// Command restorecheck is the application half of the restore
// rehearsal and of a real restore (#3340).
//
//	restorecheck verify  -dir <restic restore target> [-ledger f]
//	restorecheck reapply -data-dir <data dir> [-ledger f] [-dry-run]
//
// verify takes a tree restored from a snapshot (the data dir layout:
// db/cmdctrl.backup.sqlite, playmats/, deletions/), makes it a live
// database the way the restore runbook does, opens it with the current
// binary (which applies every pending migration), and checks it:
// integrity_check, foreign_key_check, a re-application of the deletion
// ledger, and the playmat files the database points at. It writes only
// inside -dir, so point it at a scratch directory, never at the data dir.
// It prints key=value lines and exits non-zero if any check failed.
//
// reapply is the real-restore step: after the restored database is in
// place and before the server starts, delete again every account in the
// deletion ledger that the restored database holds as live, and remove
// their playmat files. -ledger defaults to <data dir>/deletions/ledger.jsonl;
// pass the ledger restored from the LATEST snapshot when the database came
// from an older one. Safe to run twice.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, out, errw io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(errw, "usage: restorecheck verify -dir <dir> | reapply -data-dir <dir> [-ledger f] [-dry-run]")
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var err error
	switch args[0] {
	case "verify":
		fs := flag.NewFlagSet("verify", flag.ContinueOnError)
		fs.SetOutput(errw)
		dir := fs.String("dir", "", "restored tree (scratch directory)")
		ledger := fs.String("ledger", "", "ledger file (default <dir>/deletions/ledger.jsonl)")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		if *dir == "" {
			_, _ = fmt.Fprintln(errw, "restorecheck: -dir is required")
			return 2
		}
		err = verify(ctx, *dir, *ledger, out)
	case "reapply":
		fs := flag.NewFlagSet("reapply", flag.ContinueOnError)
		fs.SetOutput(errw)
		dataDir := fs.String("data-dir", os.Getenv("CMDCTRL_DATA_DIR"), "server data directory")
		ledger := fs.String("ledger", "", "ledger file (default <data dir>/deletions/ledger.jsonl)")
		dry := fs.Bool("dry-run", false, "report what would be deleted, change nothing")
		if fs.Parse(args[1:]) != nil {
			return 2
		}
		if *dataDir == "" {
			_, _ = fmt.Fprintln(errw, "restorecheck: -data-dir (or CMDCTRL_DATA_DIR) is required")
			return 2
		}
		err = reapply(ctx, *dataDir, *ledger, *dry, out)
	default:
		_, _ = fmt.Fprintf(errw, "restorecheck: unknown command %q\n", args[0])
		return 2
	}
	if err != nil {
		_, _ = fmt.Fprintln(errw, "restorecheck:", err)
		return 1
	}
	return 0
}

// verify implements the verify command; see the package comment.
func verify(ctx context.Context, dir, ledgerPath string, out io.Writer) error {
	src := filepath.Join(dir, "db", db.BackupFileName)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("restored tree has no database copy: %w", err)
	}
	live := filepath.Join(dir, "db", "cmdctrl.sqlite")
	for _, s := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(live + s); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := copyFile(src, live); err != nil {
		return err
	}
	d, err := db.Open(ctx, dir) // applies pending migrations; ErrSchemaTooNew fails here
	if err != nil {
		return fmt.Errorf("open restored database: %w", err)
	}
	defer func() { _ = d.Close() }()
	_, _ = fmt.Fprintln(out, "restorecheck: opened and migrated")

	var failures []string
	fail := func(f string, a ...any) { failures = append(failures, fmt.Sprintf(f, a...)) }

	var ic string
	if err := d.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&ic); err != nil || ic != "ok" {
		fail("integrity_check: %q %v", ic, err)
	}
	fkRows, err := d.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	violations := 0
	for fkRows.Next() {
		violations++
	}
	_ = fkRows.Close()
	if violations > 0 {
		fail("foreign_key_check: %d violations", violations)
	}

	var users, tombstones int
	if err := d.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(deleted_at) FROM users`).Scan(&users, &tombstones); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "restorecheck: users=%d tombstones=%d\n", users, tombstones)

	if ledgerPath == "" {
		ledgerPath = accountdelete.LedgerPath(dir)
	}
	entries, err := accountdelete.ReadLedger(ledgerPath)
	if err != nil {
		fail("%v", err)
	}
	rep, err := accountdelete.ReapplyLedger(ctx, accountdelete.NewSQLStore(d), entries, false)
	if err != nil {
		fail("%v", err)
	}
	_, _ = fmt.Fprintf(out, "restorecheck: ledger entries=%d reapplied=%d already_deleted=%d unknown=%d\n",
		len(entries), rep.Applied, rep.AlreadyDeleted, rep.Unknown)

	missing, err := missingPlaymats(ctx, d, filepath.Join(dir, "playmats"))
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "restorecheck: playmat_files_missing=%d\n", missing)
	if missing > 0 {
		fail("%d playmat rows have no image file in the snapshot", missing)
	}

	for _, f := range failures {
		_, _ = fmt.Fprintln(out, "restorecheck: FAIL:", f)
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d check(s) failed", len(failures))
	}
	_, _ = fmt.Fprintln(out, "restorecheck: ok")
	return nil
}

// missingPlaymats counts user_playmats rows whose image is not on disk.
func missingPlaymats(ctx context.Context, d *db.DB, playmatDir string) (int, error) {
	rows, err := d.QueryContext(ctx, `SELECT playmat_id FROM user_playmats`)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	store := playmat.NewFileStore(playmatDir)
	missing := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if !store.Exists(id) {
			missing++
		}
	}
	return missing, rows.Err()
}

// reapply implements the reapply command; see the package comment.
func reapply(ctx context.Context, dataDir, ledgerPath string, dry bool, out io.Writer) error {
	if ledgerPath == "" {
		ledgerPath = accountdelete.LedgerPath(dataDir)
	}
	entries, err := accountdelete.ReadLedger(ledgerPath)
	if err != nil {
		return err
	}
	d, err := db.Open(ctx, dataDir)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()
	rep, err := accountdelete.ReapplyLedger(ctx, accountdelete.NewSQLStore(d), entries, dry)
	if err != nil {
		return err
	}
	if !dry {
		store := playmat.NewFileStore(filepath.Join(dataDir, "playmats"))
		for _, id := range rep.PlaymatIDs {
			if err := store.Delete(id); err != nil {
				return err
			}
		}
	}
	verb := "reapplied"
	if dry {
		verb = "would_reapply"
	}
	_, _ = fmt.Fprintf(out, "restorecheck: ledger entries=%d %s=%d already_deleted=%d unknown=%d at=%s\n",
		len(entries), verb, rep.Applied, rep.AlreadyDeleted, rep.Unknown, time.Now().UTC().Format(time.RFC3339))
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	o, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(o, in); err != nil {
		_ = o.Close()
		return err
	}
	return o.Close()
}
