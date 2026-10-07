#!/usr/bin/env python3
"""Measure this project's Claude Code token usage from local transcripts.

Claude Code writes every API response's token usage to
~/.claude/projects/**/*.jsonl, subagents included. This script sums the
records whose working directory is this repo (worktrees included), per day
and per model, and merges them into a ledger file.

The ledger matters because Claude Code deletes transcripts after
`cleanupPeriodDays` (default 30). Days that have aged out of the
transcripts are kept from the ledger, so run this at least every few weeks
(or raise cleanupPeriodDays) and the history never has a hole.

The ledger holds token counts only: no prompts, no code, no file names.

    python3 scripts/cost/claude_usage.py            # update your ledger, print a summary
    python3 scripts/cost/claude_usage.py --push     # ...and publish it to the cost-ledger branch

The ledger is named after your GitHub login (from `gh`), which is how
cost_report.py matches it to a person in config.json.
"""
import argparse
import collections
import datetime as dt
import getpass
import glob
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
REPO = os.path.dirname(os.path.dirname(HERE))
FIELDS = ("requests", "input", "output", "cache_read", "cache_write_5m", "cache_write_1h")


def load_config():
    with open(os.path.join(HERE, "config.json"), encoding="utf-8") as fh:
        return json.load(fh)


def price_key(model, pricing):
    """Longest pricing key that prefixes the model ID ('claude-opus-5-5' before 'claude-opus-5')."""
    base, _, suffix = model.partition("@")
    best = None
    for k in pricing:
        if k.startswith("_") or "@" in k:
            continue
        if base.startswith(k) and (best is None or len(k) > len(best)):
            best = k
    if best and suffix == "long" and best + "@long" in pricing:
        return best + "@long"
    return best


def cost_of(model, t, pricing):
    """USD for one token bucket. Returns None for a model with no price."""
    k = price_key(model, pricing)
    if k is None:
        return None
    pi, po, pw5, pw1, pr = pricing[k]
    mult = pricing.get("_fast_mode_multiplier", 2.0) if model.endswith("@fast") else 1.0
    return mult * (t["input"] * pi + t["output"] * po + t["cache_write_5m"] * pw5
                   + t["cache_write_1h"] * pw1 + t["cache_read"] * pr) / 1e6


def scan(projects_root, repo_names):
    """Return {day: {model_bucket: Counter}} from the transcripts on this machine."""
    dir_marks = [n.replace("_", "-") for n in repo_names]
    seen = {}
    for path in glob.glob(os.path.join(projects_root, "**", "*.jsonl"), recursive=True):
        whole_dir = any(m in path for m in dir_marks)
        try:
            fh = open(path, encoding="utf-8", errors="replace")
        except OSError:
            continue
        with fh:
            for line in fh:
                if '"usage"' not in line:
                    continue
                if not whole_dir and not any(n in line for n in repo_names):
                    continue
                try:
                    d = json.loads(line)
                except ValueError:
                    continue
                if d.get("type") != "assistant":
                    continue
                if not whole_dir and not any(n in (d.get("cwd") or "") for n in repo_names):
                    continue
                m = d.get("message") or {}
                u = m.get("usage") or {}
                model = m.get("model") or ""
                if not u or not model or model.startswith("<"):
                    continue
                cc = u.get("cache_creation") or {}
                w1 = cc.get("ephemeral_1h_input_tokens") or 0
                w5 = cc.get("ephemeral_5m_input_tokens") or 0
                if not (w1 or w5):
                    w5 = u.get("cache_creation_input_tokens") or 0
                rec = {
                    "day": (d.get("timestamp") or "")[:10],
                    "model": model,
                    "fast": u.get("speed") == "fast",
                    "subagent": bool(d.get("isSidechain")) or "/subagents/" in path,
                    "input": u.get("input_tokens") or 0,
                    "output": u.get("output_tokens") or 0,
                    "cache_read": u.get("cache_read_input_tokens") or 0,
                    "cache_write_5m": w5,
                    "cache_write_1h": w1,
                }
                # One response is logged once per content block, and again in the
                # parent session when a subagent's transcript is mirrored there.
                key = (m.get("id"), d.get("requestId"))
                prev = seen.get(key)
                if prev is None or rec["output"] >= prev["output"]:
                    seen[key] = rec

    days = collections.defaultdict(lambda: collections.defaultdict(collections.Counter))
    for r in seen.values():
        bucket = r["model"]
        if bucket.startswith("claude-haiku-5-5") and r["input"] + r["cache_read"] + r["cache_write_5m"] + r["cache_write_1h"] > 100_000:
            bucket += "@long"
        if r["fast"]:
            bucket += "@fast"
        c = days[r["day"]][bucket]
        c["requests"] += 1
        for f in FIELDS[1:]:
            c[f] += r[f]
        if r["subagent"]:
            c["subagent_requests"] += 1
    return days


def merge(ledger_days, fresh_days):
    """Fresh data wins for any day it covers with at least as many requests.

    A day at the edge of the retention window can be partially deleted, so a
    fresh scan that sees FEWER requests than the ledger already holds is the
    truncated one and the ledger copy is kept.
    """
    out = dict(ledger_days)
    for day, models in fresh_days.items():
        fresh_req = sum(c["requests"] for c in models.values())
        old_req = sum(c.get("requests", 0) for c in out.get(day, {}).values())
        if fresh_req >= old_req:
            out[day] = {m: dict(c) for m, c in models.items()}
    return dict(sorted(out.items()))


def summarize(days, pricing):
    by_model = collections.defaultdict(lambda: collections.Counter())
    by_month = collections.defaultdict(float)
    total = sub = 0.0
    unpriced = set()
    for day, models in days.items():
        for model, t in models.items():
            t = {f: t.get(f, 0) for f in FIELDS + ("subagent_requests",)}
            c = cost_of(model, t, pricing)
            if c is None:
                unpriced.add(model)
                continue
            for f in FIELDS:
                by_model[model][f] += t[f]
            by_model[model]["cost"] += c
            by_month[day[:7]] += c
            total += c
            if t["requests"]:
                sub += c * t["subagent_requests"] / t["requests"]
    return by_model, dict(sorted(by_month.items())), total, sub, sorted(unpriced)


LEDGER_BRANCH = "cost-ledger"


def git(*args, stdin=None):
    return subprocess.run(("git",) + args, cwd=REPO, input=stdin, capture_output=True, text=True, check=True).stdout.strip()


def default_person():
    try:
        login = subprocess.run(("gh", "api", "user", "--jq", ".login"), capture_output=True, text=True,
                               timeout=30).stdout.strip()
    except (OSError, subprocess.TimeoutExpired):
        login = ""
    return login or getpass.getuser()


def push(ledger_path, person, attempts=3):
    """Commit the ledger as usage/claude-usage-<person>.json on the cost-ledger branch and push it.

    Pure git plumbing, so the working tree, the index and the current branch
    are never touched. The branch is created on first use.
    """
    name = f"claude-usage-{person}.json"
    for _ in range(attempts):
        try:
            git("fetch", "--quiet", "origin", f"+refs/heads/{LEDGER_BRANCH}:refs/remotes/origin/{LEDGER_BRANCH}")
            parent = git("rev-parse", f"refs/remotes/origin/{LEDGER_BRANCH}")
        except subprocess.CalledProcessError:
            parent = None
        blob = git("hash-object", "-w", ledger_path)
        usage = []
        root = []
        if parent:
            for line in git("ls-tree", f"{parent}:usage").splitlines():
                if not line.endswith("\t" + name):
                    usage.append(line)
            for line in git("ls-tree", parent).splitlines():
                if not line.endswith("\tusage"):
                    root.append(line)
        usage.append(f"100644 blob {blob}\t{name}")
        usage_tree = git("mktree", stdin="\n".join(usage) + "\n")
        root.append(f"040000 tree {usage_tree}\tusage")
        tree = git("mktree", stdin="\n".join(root) + "\n")
        if parent and tree == git("rev-parse", f"{parent}^{{tree}}"):
            print(f"{LEDGER_BRANCH}: {name} unchanged")
            return
        cmd = ["commit-tree", tree, "-m", f"chore(cost): usage ledger for {person}"]
        if parent:
            cmd[2:2] = ["-p", parent]
        commit = git(*cmd)
        try:
            git("push", "--quiet", "origin", f"{commit}:refs/heads/{LEDGER_BRANCH}")
            print(f"pushed {name} to {LEDGER_BRANCH}")
            return
        except subprocess.CalledProcessError as e:
            # Someone else pushed their ledger in between: refetch and rebuild on top.
            last = e.stderr
    sys.exit(f"push to {LEDGER_BRANCH} failed: {last}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--person", default=None, help="ledger name (default: your GitHub login, else $USER)")
    ap.add_argument("--projects", default=os.path.expanduser("~/.claude/projects"), help="Claude Code transcript root")
    ap.add_argument("--ledger-dir", default=os.path.join(REPO, "data", "cost"), help="where ledgers live (gitignored)")
    ap.add_argument("--no-write", action="store_true", help="print the summary without touching the ledger")
    ap.add_argument("--push", action="store_true", help=f"publish the ledger to the {LEDGER_BRANCH} branch")
    args = ap.parse_args()

    cfg = load_config()
    pricing = cfg["pricing_per_mtok"]
    person = args.person or default_person()
    repo_names = [os.path.basename(REPO), "cmd_and_ctrl"]
    ledger_path = os.path.join(args.ledger_dir, f"claude-usage-{person}.json")

    ledger = {"schema": 1, "person": person, "days": {}}
    if os.path.exists(ledger_path):
        with open(ledger_path, encoding="utf-8") as fh:
            ledger = json.load(fh)

    fresh = scan(args.projects, sorted(set(repo_names)))
    ledger["days"] = merge(ledger.get("days", {}), fresh)
    ledger["updated"] = dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")
    ledger["transcripts_cover_from"] = min(fresh) if fresh else None

    if not args.no_write:
        os.makedirs(args.ledger_dir, exist_ok=True)
        with open(ledger_path, "w", encoding="utf-8") as fh:
            json.dump(ledger, fh, indent=1, sort_keys=True)
            fh.write("\n")

    by_model, by_month, total, sub, unpriced = summarize(ledger["days"], pricing)
    days = list(ledger["days"])
    print(f"Claude usage for {person}: {days[0] if days else '-'} .. {days[-1] if days else '-'}"
          f"  ({len(days)} active days)")
    print(f"{'model':28s} {'requests':>9s} {'output':>13s} {'cache read':>16s} {'API $':>11s}")
    for model, c in sorted(by_model.items(), key=lambda kv: -kv[1]["cost"]):
        print(f"{model:28s} {c['requests']:9,d} {c['output']:13,d} {c['cache_read']:16,d} {c['cost']:11,.2f}")
    print(f"{'total':28s} {'':9s} {'':13s} {'':16s} {total:11,.2f}  (subagents ${sub:,.0f})")
    print("by month: " + ", ".join(f"{m} ${v:,.0f}" for m, v in by_month.items()))
    if unpriced:
        print("no price in config.json for: " + ", ".join(unpriced), file=sys.stderr)
    if not args.no_write:
        print(f"ledger: {os.path.relpath(ledger_path, REPO)}")
    if args.push and not args.no_write:
        push(ledger_path, person)
    if fresh and days and min(fresh) == days[0]:
        print(f"note: the ledger starts where this machine's transcripts do ({days[0]}). Anything older was"
              " already deleted by Claude Code (cleanupPeriodDays) and cannot be measured.", file=sys.stderr)


if __name__ == "__main__":
    main()
