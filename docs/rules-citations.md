# Rules citations

The rule text and the edition pin are summarised in [AGENTS.md](../AGENTS.md) §6. This file has the procedure.

`CR NNN.Nx` in code, tests and docs means the **Magic: The Gathering
Comprehensive Rules effective September 25, 2026**, from
[magic.wizards.com/en/rules](https://magic.wizards.com/en/rules) (the TXT
download is `MagicCompRules 20260925.txt`; its text says "effective as of
September 25, 2026"). Check a number against that text before you write it. Do not
cite from memory: rule numbers move between editions. The June 2025 edition
re-sorted every keyword action in 701, so discard went from 701.8 to 701.9 and
reveal from 701.16 to 701.20. The 2026 edition added a new 310.8, which moved
the battle protector rules to 310.9. The September 2026 edition added a new
506.6 ("attacks a player alone"), which moved "had to attack" to 506.7 and the
combat timing rules for spells (old 506.7–506.7g) to 506.8–506.8g.

To move the pin to a newer edition, do it in one PR of its own. Download the
new TXT. For every section the tree cites, compare the rule's text in the two
editions, and renumber by matching content, never by adding to the number.
Then update the date and file name above.

A one-line grep does not find every citation. Comments wrap, so "CR" often
ends one line and the number starts the next. List the sections with a search
that crosses line breaks:

```sh
rg -U -o -N --no-filename 'CR\s*(//|\*|#)?\s*[0-9]{3}\.[0-9]+[a-z]?' . < /dev/null \
  | grep -oE '[0-9]{3}\.[0-9]+[a-z]?' | sort -u
```

Some citations have no "CR" in front at all: the second number in
"CR 305.1, 116.2a", and rule tables in comments such as the state-based
action list in `mutations.go`. Once you know which numbers moved, search for
each old number on its own too (`git grep -nw "701\.19"`), then read each hit.
