package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eddie Brock // Venom, Lethal Protector — a double-faced Legendary
// Creature (Marvel's Spider-Man, #4460), one file for both faces:
//
//	Eddie Brock {2}{B}, Legendary Creature — Human Hero Villain, 3/3
//	  "When Eddie Brock enters, return target creature card with mana
//	   value 1 or less from your graveyard to the battlefield.
//	   {3}{B}{R}{G}: Transform Eddie Brock. Activate only as a sorcery."
//
//	Venom, Lethal Protector {3}{B}{R}{G}, Legendary Creature — Symbiote
//	Hero Villain, 5/5 (key "<oracle>#1", Scryfall's modal_dfc layout)
//	  "Menace, trample, haste
//	   Whenever Venom attacks, you may sacrifice another creature. If you
//	   do, draw X cards, then you may put a permanent card with mana
//	   value X or less from your hand onto the battlefield, where X is
//	   the sacrificed creature's mana value."
//
// Scryfall lays the card out as a modal double-faced card, and every
// double-faced card can transform (CR 712.9), so the activation turns
// the permanent over in place (Bruce Banner's TransformThis).
//
// The enters trigger targets in the controller's graveyard; the
// returned creature arrives under its owner's (and so their) control.
// Venom's "you may" is the sacrifice prompt itself (maySacrificeOne,
// Ruthless Technomancer's shape): "if you do" is the sacrifice's own
// answer, X is the creature's last-known mana value (CR 608.2h; a
// token that is no copy is 0), the draw comes first, and the put is read
// from the hand as it stands after the draw. Only permanent cards are
// offered (CR 110.4).
//
// No simplification.
func init() {
	const oracle = "fe01c96b-eb78-4477-bec2-944b2a28c778"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Eddie Brock",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Eddie Brock — return target creature card with mana value 1 or less from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
				}),
				TargetCardInGraveyard("target creature card with mana value 1 or less in your graveyard", Creature(), YouOwn(), ManaValueLE(1))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}{R}{G}: Transform Eddie Brock. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{B}{R}{G}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Venom, Lethal Protector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Venom, Lethal Protector — you may sacrifice another creature; if you do, draw X cards, then you may put a permanent card with mana value X or less from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item,
						"Venom, Lethal Protector — sacrifice another creature to draw cards equal to its mana value?",
						Creature(), true, venomLethalProtectorSacrificed)
				}),
		},
	})
}

// venomLethalProtectorSacrificed is "If you do, draw X cards, then you
// may put a permanent card with mana value X or less from your hand onto
// the battlefield", once the creature has really gone, keyed (#4290).
var venomLethalProtectorSacrificed = RegisterSacrificedOne("prompt-run/venom-lethal-protector-sacrificed", func(ctx *Context, sacrificed uuid.UUID) error {
	x := departedManaValue(ctx.Game, sacrificed)
	if err := (DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
		return err
	}
	return PutFromHandOntoBattlefield{
		Match:    ManaValueLE(x),
		Optional: true,
		Label:    "Venom, Lethal Protector — you may put a permanent card with mana value X or less from your hand onto the battlefield",
	}.Apply(ctx)
})

// departedManaValue is the mana value of the permanent object `cardID`
// last was on the battlefield (CR 608.2h), zero when the engine has no
// record of it.
func departedManaValue(g *game.Game, cardID uuid.UUID) int {
	ref, ok := g.PermanentRefForEffect(cardID)
	if !ok {
		return 0
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.ManaValue < 0 {
		return 0
	}
	return info.ManaValue
}
