package effects

// hexproof_bodies.go is the deliberate exception to one-file-per-card.
// Gladecover Scout and Slippery Bogle are the same card twice — a
// one-mana 1/1 with hexproof and no other text — and they are the
// cheap voltron shells the archetype is built on: a creature your
// opponent cannot answer with removal, which makes every pump spell
// and (from S33) every Aura you put on it a permanent investment
// rather than a two-for-one waiting to happen.
//
// Two `Register` calls in one `init()` rather than two files because
// there is genuinely nothing to say about either of them
// individually, and a file whose entire content is a `PrintedKeywords`
// slot is worse for `git blame` than a shared file with a reason at
// the top. A third one-mana hexproof body joins this file; anything
// with a second line of text gets its own.
//
// Both are fully live. `hexproof` is enforced at the targeting choke
// point (game/targets.go via CanBeTargetedBy) as of S23, so these
// really do refuse an opponent's Doom Blade while still accepting
// your own Blossoming Defense — that asymmetry is CR 702.11b and is
// the whole reason the archetype works.
//
// No simplifications.
func init() {
	// Gladecover Scout — Creature — Elf Scout, {G}, 1/1.
	Register(Spec{
		OracleID:        "385c1208-bfea-44e2-b236-4f38bc90db9f",
		Name:            "Gladecover Scout",
		PrintedKeywords: []string{"hexproof"},
	})

	// Slippery Bogle — Creature — Beast, {G/U}, 1/1. The hybrid cost
	// is printed data the deck importer carries; nothing here needs
	// to know about it.
	Register(Spec{
		OracleID:        "f64878c1-1fba-44f7-a24f-d24bef2e03ae",
		Name:            "Slippery Bogle",
		PrintedKeywords: []string{"hexproof"},
	})
}
