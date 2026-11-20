package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonstone Harbinger — Creature — Bat Warrior {2}{B}, 1/3
// (Bloomburrow, #4467):
//
//	"Flying, deathtouch
//	 Whenever you gain or lose life during your turn, Bats you control
//	 get +1/+0 and gain deathtouch until end of turn. This ability
//	 triggers only once each turn."
//
// The trigger watches both life-change events and damage: a gain is a
// positive EventChangeLife, a loss is a negative one or damage that cost
// life (s22PlayerLostLife, which Vengeful Warchief's first loss reads
// too). It fires only while its controller is the active player, and
// "only once each turn" is the per-object trigger tally (Exemplar of
// Light's shape): a second gain or loss that turn finds the first
// trigger already counted. The Bats are chosen as it resolves and keep
// the bonus for the turn (CR 611.2c), so a Bat that enters afterwards
// does not get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcae4f9d-8f81-42f9-93a3-82a482851bdd",
		Name:            "Moonstone Harbinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			Key:       moonstoneHarbingerLabel,
			AppliesTo: moonstoneHarbingerLifeChangedOnYourTurn,
			Effect:    moonstoneHarbingerPumpBats,
		}},
	})
}

const moonstoneHarbingerLabel = "Moonstone Harbinger — Bats you control get +1/+0 and gain deathtouch until end of turn"

// moonstoneHarbingerLifeChangedOnYourTurn is the trigger condition: you
// gained or lost life, it is your turn, and the ability has not
// triggered yet this turn.
func moonstoneHarbingerLifeChangedOnYourTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	you := source.Controller
	if !effects.IsYourTurn(g, you) {
		return false
	}
	_, lost := effects.S22PlayerLostLife(ev, you, g)
	if !lost && !effects.YouGainedLife(ev, source, lki, g) {
		return false
	}
	return !effects.B11TriggeredThisTurn(g, source.InstanceID, moonstoneHarbingerLabel)
}

func moonstoneHarbingerPumpBats(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	bats := effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Bat"))
	if err := (effects.BoostUntilEOT{Match: bats, Power: 1, Label: "Moonstone Harbinger — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Match: bats, Keywords: []string{"deathtouch"}, Label: "Moonstone Harbinger — deathtouch"}.Apply(ctx)
}
