# Card workers: the #5906 split

Before you build, run `git fetch origin develop` and check whether `server/internal/cards/effects/shards.txt` exists on origin/develop.

- **It does not exist:** the split has not landed. Follow BRIEF.md as written; card files go in `server/internal/cards/effects/`.
- **It exists:** the split has landed. Read the "This takes effect when #5906's split lands" section of docs/adding-cards.md and follow it.
  - Write each card and test in the flat directory as BRIEF.md says, then run `scripts/split-effects.sh --place <files>` from the repository root before your checks. It moves each file to its shard or test package.
  - Card files call core helpers qualified (`effects.X`).
  - Run `go test ./internal/cards/effects/...` and `TestEffectsTreeLayout`.

Check again before each new batch: the split can land between your batches. A batch branch made fresh from develop after the split must use the new layout.
