package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ishgard, the Holy See // Faith & Grief — Land — Town, with a Sorcery
// — Adventure half {3}{W}{W} (oracle 4f4358cb…, #2176):
//
//	Ishgard, the Holy See
//	  "This land enters tapped.
//	   {T}: Add {W}."
//	Faith & Grief
//	  "Return up to two target artifact and/or enchantment cards from
//	   your graveyard to your hand. (Then exile this card. You may
//	   play the land later from exile.)"
//
// A Town land with an Adventure half; see Jidoor for the land-play
// lifecycle. Targets are declared as a 0-2 clause, so a card that
// left the graveyard in response is skipped and the other still
// returns (CR 608.2b).
//
// No simplification.
const ishgardOracleID = "4f4358cb-59df-46d9-be27-69929f5a615c"

func init() {
	Register(Spec{
		OracleID:     ishgardOracleID,
		Name:         "Ishgard, the Holy See",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
	Register(Spec{
		OracleID:     ishgardOracleID + "#1",
		Name:         "Faith & Grief",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to two target artifact and/or enchantment cards from your graveyard",
			YouOwn(), Or(Artifact(), Enchantment())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
