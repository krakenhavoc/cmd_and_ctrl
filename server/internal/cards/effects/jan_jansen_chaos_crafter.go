package effects

// Jan Jansen, Chaos Crafter — Legendary Creature — Gnome Artificer
// {R}{W}{B}, 3/3 (EDHREC rank 4111):
//
//	"Haste
//	 {T}, Sacrifice an artifact creature: Create two Treasure tokens.
//	 {T}, Sacrifice a noncreature artifact: Create two 1/1 colorless
//	 Construct artifact creature tokens."
//
// A three-mana Mardu artifact engine that launders artifacts into
// other artifacts. The two abilities are deliberately complementary:
// the first eats a body and pays in mana, the second eats a rock and
// pays in bodies, and a Construct made by the second is fuel for the
// first. With any untapper the loop is a Treasure a turn; with a
// sacrifice payoff on the board it is a Treasure and two death
// triggers.
//
// THE TWO COSTS PARTITION THE ARTIFACTS, exactly as printed. "An
// artifact creature" and "a noncreature artifact" are disjoint, which
// is why they are two abilities and not one with a choice: a Treasure
// can only ever pay the second, a Construct token only the first.
// Both are read as EFFECTIVE types, so an animated Treasure (Tempered
// Steel is not enough, but an Ensoul Artifact is) pays the first cost
// and stops paying the second, which is correct.
//
// BOTH TAP, so only one fires per turn without help — that is the
// governor on the engine and it is a cost, not an effect.
//
// The sacrifice is a COST paid at announce (CR 601.2h): the artifact
// is already in the graveyard when the ability goes on the stack, so
// a Mayhem Devil or a Bastion of Remembrance triggers ABOVE it and
// resolves first, and countering the ability does not give the
// artifact back. Jan himself is a Gnome Artificer and not an
// artifact, so he can never eat himself.
//
// Haste rides PrintedKeywords, which matters here: the abilities have
// a tap symbol, so without haste Jan would do nothing the turn he
// lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb491d8a-2e9f-46fe-9590-44c3b4a25f1b",
		Name:            "Jan Jansen, Chaos Crafter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{
			{
				Label: "{T}, Sacrifice an artifact creature: Create two Treasure tokens.",
				Cost:  Plus(TapCost(), SacrificeN(1, "an artifact creature", Artifact(), Creature())),
				Effect: Do(CreateToken{
					Template: TreasureToken(),
					N:        2,
				}),
			},
			{
				Label: "{T}, Sacrifice a noncreature artifact: Create two 1/1 colorless Construct artifact creature tokens.",
				Cost:  Plus(TapCost(), SacrificeN(1, "a noncreature artifact", Artifact(), Noncreature())),
				Effect: Do(CreateToken{
					Template: TokenCard("1/1 colorless Construct artifact"),
					N:        2,
				}),
			},
		},
	})
}
