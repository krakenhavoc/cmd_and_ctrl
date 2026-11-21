package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const IshgardOracleID = "4f4358cb-59df-46d9-be27-69929f5a615c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     IshgardOracleID,
		Name:         "Ishgard, the Holy See",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     IshgardOracleID + "#1",
		Name:         "Faith & Grief",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to two target artifact and/or enchantment cards from your graveyard",
			effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
