package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/norfrt6-lab/go-dev-cli/internal/database"
	"github.com/norfrt6-lab/go-dev-cli/internal/model"
	"github.com/norfrt6-lab/go-dev-cli/internal/service"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Aggregate, search, and tail log files",
	Long:  "Tail log files, full-text search stored entries, and manage log retention.",
}

var logsTailCmd = &cobra.Command{
	Use:   "tail <file>",
	Short: "Tail a log file (like tail -f)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")
		svc := service.NewLogService(getDB())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			cancel()
		}()

		count := 0
		return svc.TailFile(ctx, args[0], follow, func(entry *model.LogEntry) {
			count++
			if !follow && lines > 0 && count > lines {
				return
			}
			printLogEntry(entry)
		})
	},
}

var logsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search across stored log entries",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		query := strings.Join(args, " ")

		svc := service.NewLogService(getDB())
		entries, err := svc.Search(query, limit)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("No log entries found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIME\tLEVEL\tSOURCE\tMESSAGE")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				e.Timestamp.Format("15:04:05"),
				levelTag(e.Level),
				e.Source,
				truncate(e.Message, 80),
			)
		}
		return w.Flush()
	},
}

var logsQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query stored log entries with filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		level, _ := cmd.Flags().GetString("level")
		limit, _ := cmd.Flags().GetInt("limit")
		sinceStr, _ := cmd.Flags().GetString("since")

		q := database.LogQuery{
			Source: source,
			Limit:  limit,
		}

		if level != "" {
			q.Level = model.LogLevel(level)
		}

		if sinceStr != "" {
			dur, err := time.ParseDuration(sinceStr)
			if err != nil {
				return fmt.Errorf("invalid --since duration: %w", err)
			}
			t := time.Now().Add(-dur)
			q.Since = &t
		}

		svc := service.NewLogService(getDB())
		entries, err := svc.Query(q)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Println("No log entries found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIME\tLEVEL\tSOURCE\tMESSAGE")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				e.Timestamp.Format("2006-01-02 15:04:05"),
				levelTag(e.Level),
				e.Source,
				truncate(e.Message, 80),
			)
		}
		return w.Flush()
	},
}

var logsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear stored log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		retentionDays, _ := cmd.Flags().GetInt("older-than")

		svc := service.NewLogService(getDB())

		if retentionDays > 0 {
			deleted, err := svc.Cleanup(retentionDays)
			if err != nil {
				return err
			}
			fmt.Printf("Deleted %d log entries older than %d days.\n", deleted, retentionDays)
			return nil
		}

		deleted, err := svc.Clear(source)
		if err != nil {
			return err
		}

		if source != "" {
			fmt.Printf("Deleted %d log entries from '%s'.\n", deleted, source)
		} else {
			fmt.Printf("Deleted %d log entries.\n", deleted)
		}
		return nil
	},
}

var logsSourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List all log sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := service.NewLogService(getDB())
		sources, err := svc.Sources()
		if err != nil {
			return err
		}

		if len(sources) == 0 {
			fmt.Println("No log sources found.")
			return nil
		}

		for _, s := range sources {
			fmt.Println(s)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsTailCmd.Flags().BoolP("follow", "f", false, "follow file for new entries")
	logsTailCmd.Flags().IntP("lines", "n", 0, "number of lines to show (0 = all)")
	logsCmd.AddCommand(logsTailCmd)

	logsSearchCmd.Flags().IntP("limit", "l", 100, "max results")
	logsCmd.AddCommand(logsSearchCmd)

	logsQueryCmd.Flags().StringP("source", "s", "", "filter by source")
	logsQueryCmd.Flags().String("level", "", "filter by level (debug, info, warn, error)")
	logsQueryCmd.Flags().IntP("limit", "l", 50, "max results")
	logsQueryCmd.Flags().String("since", "", "show entries since duration (e.g. 1h, 30m)")
	logsCmd.AddCommand(logsQueryCmd)

	logsClearCmd.Flags().StringP("source", "s", "", "clear entries for specific source only")
	logsClearCmd.Flags().Int("older-than", 0, "clear entries older than N days")
	logsCmd.AddCommand(logsClearCmd)

	logsCmd.AddCommand(logsSourcesCmd)
}

func printLogEntry(entry *model.LogEntry) {
	fmt.Printf("%s %s %s\n",
		entry.Timestamp.Format("15:04:05"),
		levelTag(entry.Level),
		entry.Message,
	)
}

func levelTag(level model.LogLevel) string {
	switch level {
	case model.LogDebug:
		return "[DBG]"
	case model.LogInfo:
		return "[INF]"
	case model.LogWarn:
		return "[WRN]"
	case model.LogError:
		return "[ERR]"
	default:
		return "[???]"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
