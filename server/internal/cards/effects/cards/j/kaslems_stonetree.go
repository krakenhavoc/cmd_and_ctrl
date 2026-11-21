package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaslem's Stonetree // Kaslem's Strider — a transforming artifact
// (#2124, ADR 0137):
//
//	Kaslem's Stonetree — Artifact {2}{G}
//	  "When this artifact enters, look at the top six cards of your
//	   library. You may put a land card from among them onto the
//	   battlefield tapped. Put the rest on the bottom in a random order.
//	   Craft with Cave {5}{G}"
//	Kaslem's Strider — Artifact Creature — Golem, 5/5
//
// Armored Skyhunter's look-at-six with a land filter and a tapped
// entry; only the controller sees the six. Craft with Cave is the first
// subtype material (CR 702.167b): a Cave you control or a Cave card in
// your graveyard, whatever its card type. The Strider is vanilla, so it
// needs no entry of its own.
//
// No simplification.
const KaslemsStonetreeOracleID = "1ac3e4bc-1678-4280-a071-3dbc8ef4a2bf"

func init() {
	effects.Register(effects.Spec{
		OracleID:     KaslemsStonetreeOracleID,
		Name:         "Kaslem's Stonetree",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kaslem's Stonetree — look at the top six; you may put a land onto the battlefield tapped", kaslemsStonetreeLook),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with Cave {5}{G}", "{5}{G}", effects.CraftWithSubtype("Cave")),
		},
	})
}

func kaslemsStonetreeLook(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return effects.PutFromLibraryOntoBattlefield{
		Player:   controller,
		Cards:    g.LookAtTopOfLibraryForEffect(controller, 6),
		Match:    effects.Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Kaslem's Stonetree — you may put a land card from among them onto the battlefield tapped",
		ThenTail: effects.PutRestOnBottomInRandomOrderTail,
	}.Apply(effects.NewContext(g, item))
}
