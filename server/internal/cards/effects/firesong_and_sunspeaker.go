package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firesong and Sunspeaker — Legendary Creature — Minotaur Cleric
// {4}{R}{W}, 4/6:
//
//	"Red instant and sorcery spells you control have lifelink.
//	 Whenever a white instant or sorcery spell causes you to gain life,
//	 Firesong and Sunspeaker deals 3 damage to target creature or player."
//
// The static is Pestilent Spirit's layer-6 spell grant, so the lifelink is
// the spell's own as it deals damage from the stack (CR 702.15). The
// trigger reads a life gain (a positive EventChangeLife naming you) whose
// source is a white instant or sorcery SPELL on the stack, any player's
// (the gain is what makes it yours), whether the spell's own text gains the
// life or its lifelink does. A red and white spell of yours therefore
// gains the life with its granted lifelink and triggers this. The target is
// a creature or a player (not a planeswalker), chosen as the trigger is put
// on the stack.
//
// No simplification.
func init() {
	hit := On(game.EventChangeLife, whiteInstantOrSorcerySpellCausedYouToGainLife,
		"Firesong and Sunspeaker — 3 damage to target creature or player",
		damageToFirstTargetEffect(3))
	hit.Targets = targetCreatureOrPlayer("target creature or player")
	Register(Spec{
		OracleID:     "fffaa634-42a0-4038-b027-24f28754fec7",
		Name:         "Firesong and Sunspeaker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(SpellsYouControlHave(And(Or(Instant(), Sorcery()), OfColor("R")), "lifelink"),
				"Red instant and sorcery spells you control have lifelink."),
		},
		Triggered: []game.TriggeredAbility{hit},
	})
}

// whiteInstantOrSorcerySpellCausedYouToGainLife is the trigger test:
// the life change is a gain by the source's controller, caused by a white
// instant or sorcery spell that is on the stack.
func whiteInstantOrSorcerySpellCausedYouToGainLife(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventChangeLife || ev.Amount <= 0 || ev.Target != source.Controller || ev.Source == uuid.Nil {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.Source); z == nil || z.Kind != game.ZoneStack {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.Source)
	return ok && spell.HasColorWord("W") && (spell.IsInstant() || spell.IsSorcery())
}

// targetCreatureOrPlayer is "target creature or player" (no planeswalker,
// battle): Firesong and Sunspeaker.
func targetCreatureOrPlayer(label string) *game.TargetSpec {
	spec := TargetAny()
	spec.Label = label
	spec.CardOK = func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
		return c.IsCreature()
	}
	return spec
}

// damageToFirstTargetEffect is "deals n damage to <the target>" as a
// trigger body, the source being the trigger's own.
func damageToFirstTargetEffect(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return damageToFirstTarget(n)(item, ctx)
	}
}
