package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Pyrulea // Gargantuan Slabhorn — Battle — Siege, defense 4,
// for {G}{U} (March of the Machine):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, scry 3, then reveal the top card of your
//	 library. If it's a land or double-faced card, draw a card."
//
// Back face, Creature — Beast 4/4:
//
//	"Trample, ward {2}
//	 Other transformed permanents you control have trample and ward {2}."
//
// The protector prompt, defense counters and the transformed cast are engine
// behaviour keyed on the battle type (battle.go). The Siege's trigger scries
// first and reveals once the scry is done (Scry.ThenTail), so the card
// revealed is the one the scry left on top; it stays on top, and a land or a
// double-faced card is drawn (it is the card drawn, nothing else being able
// to move it in between).
//
// The back face registers under "<oracle_id>#1". Its grant is to OTHER
// permanents that are transformed (CR 701.27g, TransformedPermanent): a
// permanent with its front face up is not one, and the Slabhorn, itself
// transformed, is excluded by "other". Trample is a keyword grant, and ward is
// a triggered ability, so it is WardGranted over the same recipients.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     invasionOfPyruleaOracleID,
		Name:         "Invasion of Pyrulea",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3, Draws: 1},
		Battle: &effects.BattleSpec{
			Defense: 4,
			Subtype: effects.BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Pyrulea — scry 3, then reveal the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{Player: ctx.Controller(), N: 3, ThenTail: ctx.ItemTail(invasionOfPyruleaReveal)}.Apply(ctx)
				}),
			effects.DefeatedTrigger("Invasion of Pyrulea — defeated: exile it, then cast Gargantuan Slabhorn", effects.SiegeDefeated()),
		},
	})

	otherTransformed := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.Controller == source.Controller && effects.TransformedPermanent(*target)
	}
	effects.Register(effects.Spec{
		OracleID:        invasionOfPyruleaOracleID + "#1",
		Name:            "Gargantuan Slabhorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "ward"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(otherTransformed, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Gargantuan Slabhorn — ward {2}"),
			effects.WardGranted(effects.WardMana("{2}"), "Gargantuan Slabhorn — other transformed permanents you control have ward {2}", otherTransformed),
		},
	})
}

// invasionOfPyruleaReveal is "reveal the top card of your library. If it's a
// land or double-faced card, draw a card", once the scry is done. An on-disk
// key.
var invasionOfPyruleaReveal = effects.CardTail("invasion-of-pyrulea/reveal-then-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{Player: ctx.Controller(), N: 1, Reason: "Invasion of Pyrulea", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(revealed[0])
	if !ok || !(c.IsLand() || game.IsDoubleFacedForEffect(c)) {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

// invasionOfPyruleaOracleID is shared with the back face, which registers
// under it plus "#1".
const invasionOfPyruleaOracleID = "323437f5-cc6c-4579-b591-499476dffb3c"
