package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buster Sword — Artifact — Equipment {3}:
//
//	"Equipped creature gets +3/+2.
//	 Whenever equipped creature deals combat damage to a player, draw
//	 a card, then you may cast a spell from your hand with mana value
//	 less than or equal to that damage without paying its mana cost.
//	 Equip {2}"
//
// SIMPLIFICATION, DECLARED: the draw happens; the free cast after it
// does not. Casting a spell from HAND without paying its cost, chosen
// interactively and gated by a mana-value ceiling computed at
// resolution, has no engine hook to build on. game.CastPermission —
// the machinery every other "you may cast without paying" card in the
// catalog rides (Aloe Alchemist, Rabble Rousing, Maelstrom Colossus) —
// says so explicitly in its own doc comment: "Zone is where the cast
// comes FROM: exile, a graveyard, or a library. Never hand (CR 601.2
// already allows it)". That machinery exists to grant permission to
// cast from a zone CR 601.2 doesn't already reach; it was never meant
// to also suppress a hand-cast's cost, and repurposing it for that
// would be inventing a new engine verb inside a card file rather than
// declaring the gap. Until a real "cast this hand card for free"
// primitive exists, the second half of the trigger is left out —
// weaker than printed, never stronger.
//
// The draw itself needs nothing new: attachedCreatureDealtCombatDamageToPlayer
// is the Swords' own condition, reused verbatim.
func init() {
	Register(Spec{
		OracleID:     "5e060d58-4d6e-425c-b7d4-727669fcce5b",
		Name:         "Buster Sword",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"After the draw, you don't get to cast a free spell from hand — that half of the trigger isn't implemented.",
		},
		Static: []game.StaticAbility{
			PumpAttached(3, 2),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key:    "Buster Sword — draw a card",
			Effect: Do(DrawCards{N: 1}),
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
