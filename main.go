package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	mcputils "sonarqube-mcp/internal/helpers"
	mcpcli "sonarqube-mcp/internal/mcpcli"
	mcpserver "sonarqube-mcp/internal/mcpserver"
)

func logHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		v := mcputils.GetVerbosity()

		sessionID := r.Header.Get("Mcp-Session-Id")
		ctx := mcputils.WithSessionID(r.Context(), sessionID)
		r = r.WithContext(ctx)

		if v >= 2 {
			var bodyLog string
			if r.Body != nil {
				body, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				bodyLog = string(body)
			}
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			fmt.Fprintf(os.Stderr, "%s [http] sid=%s %s %s body=%s\n", start.Format(time.RFC3339), sid, r.Method, r.URL.RequestURI(), truncate(bodyLog, 300))
		}

		lw := &loggingWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)

		if v >= 1 {
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			fmt.Fprintf(os.Stderr, "%s [http] sid=%s %d %s %s (%s)\n", time.Now().Format(time.RFC3339), sid, lw.status, r.Method, r.URL.RequestURI(), time.Since(start).Round(time.Millisecond))
		}
	})
}

type loggingWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggingWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: sonarqube-mcp [OPTIONS] [TOOL_NAME] [JSON_ARGS]

Options:
  -t, --transport <stdio|http|cli>  Transport mode (default "stdio")
                                     cli mode: invoke a tool directly from the command line
  --port <number>                    HTTP server port (default 8080)
  -v, --verbose <0-10>              Request logging verbosity level
                                     0=silent, 1=access log, 2+=method+URL,
                                     3+=query, 5+=headers, 7+=body, 9+=pretty JSON
  -h, --help                        Show this help message

Environment:
  SONARQUBE_URL           SonarQube Server URL (required unless SONARQUBE_ORG is set)
  SONARQUBE_TOKEN         SonarQube user token (required for stdio mode)
  SONARQUBE_ORG           SonarQube Cloud organization key (implies Cloud mode)
  SONARQUBE_PROJECT_KEY   Default project key for tools that need one
  SONARQUBE_TOOLSETS      Comma-separated tool categories to enable
  SONARQUBE_READ_ONLY     Set to "true" to expose only read-only tools

CLI Mode:
  sonarqube-mcp -t cli list                   List all available tools
  sonarqube-mcp -t cli <tool-name> [OPTIONS]   Invoke a tool with GNU-style options
                                               Use --help for tool-specific help
`)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	if s == "" {
		return "(empty)"
	}
	return s
}

func main() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = usage

	transport := flag.String("t", "stdio", "Transport mode: stdio, http, or cli")
	flag.StringVar(transport, "transport", "stdio", "Transport mode: stdio, http, or cli")
	port := flag.Int("port", 8080, "HTTP server port (only used when transport=http)")
	verbose := flag.Int("v", 0, "Request logging verbosity level (0-10)")
	flag.IntVar(verbose, "verbose", 0, "Request logging verbosity level (0-10)")
	flag.Parse()

	mcputils.SetVerbosity(*verbose)

	sonarQubeURL := mcputils.GetSonarQubeURL()
	if sonarQubeURL == "" {
		fmt.Fprintln(os.Stderr, "Error: SONARQUBE_URL is required (or set SONARQUBE_ORG for Cloud)")
		os.Exit(1)
	}
	if mcputils.IsCloud() {
		fmt.Fprintf(os.Stderr, "SonarQube Cloud | org=%s | url=%s\n", mcputils.GetSonarQubeOrg(), sonarQubeURL)
	} else {
		fmt.Fprintf(os.Stderr, "SonarQube Server | url=%s\n", sonarQubeURL)
	}

	switch *transport {
	case "cli":
		if flag.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "Usage: %s -t cli <tool-name> [OPTIONS]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "       %s -t cli list\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "\nUse -t cli <tool-name> --help for tool-specific help.\n")
			os.Exit(1)
		}
		subcmd := flag.Arg(0)
		switch subcmd {
		case "list":
			mcpcli.ListTools()
		default:
			if subcmd == "--help" || subcmd == "-h" {
				fmt.Fprintf(os.Stderr, "Usage: %s -t cli <tool-name> [OPTIONS]\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "       %s -t cli list\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "\nUse -t cli <tool-name> --help for tool-specific help.\n")
				os.Exit(0)
			}
			toolName := subcmd
			args := flag.Args()[1:]
			if err := mcpcli.Call(os.Args[0], toolName, args); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}

	case "http":
		s := mcpserver.NewMCPServer()

		mcpServer := server.NewStreamableHTTPServer(s,
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				return mcputils.WithHTTPHeaders(ctx, r.Header)
			}),
		)

		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpServer)

		addr := fmt.Sprintf(":%d", *port)
		httpServer := &http.Server{Addr: addr, Handler: logHTTP(mux)}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		go func() {
			fmt.Fprintf(os.Stderr, "MCP server listening on %s/mcp\n", addr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "Shutting down...")
		if err := httpServer.Shutdown(context.WithoutCancel(ctx)); err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server shutdown error: %v\n", err)
		}

	case "stdio":
		s := mcpserver.NewMCPServer()

		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown transport: %s (use 'stdio', 'http', or 'cli')\n", *transport)
		os.Exit(1)
	}
}
