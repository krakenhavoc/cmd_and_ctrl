package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const kaslemsStonetreeOracleID = "1ac3e4bc-1678-4280-a071-3dbc8ef4a2bf"

func init() {
	Register(Spec{
		OracleID:     kaslemsStonetreeOracleID,
		Name:         "Kaslem's Stonetree",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kaslem's Stonetree — look at the top six; you may put a land onto the battlefield tapped", kaslemsStonetreeLook),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with Cave {5}{G}", "{5}{G}", CraftWithSubtype("Cave")),
		},
	})
}

func kaslemsStonetreeLook(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return PutFromLibraryOntoBattlefield{
		Player:   controller,
		Cards:    g.LookAtTopOfLibraryForEffect(controller, 6),
		Match:    Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Kaslem's Stonetree — you may put a land card from among them onto the battlefield tapped",
		Then:     PutRestOnBottomInRandomOrder,
	}.Apply(NewContext(g, item))
}
