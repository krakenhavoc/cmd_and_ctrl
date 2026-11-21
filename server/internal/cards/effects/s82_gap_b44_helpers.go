package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB44NontokenCreaturesDiedUnder is how many nontoken creatures died
// under `player`'s control this turn (Vile Redeemer): each creature that
// left the battlefield for a graveyard this turn, a creature and a card as
// it last existed on the battlefield (CR 603.10a) and under that player's
// control as it left (CR 700.4), read off this turn's event log. A token
// that died is not counted, whatever it was.
func gapB44NontokenCreaturesDiedUnder(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		dead, ok := diedCreature(ev, g)
		if !ok || IsToken(dead) || leftUnderControlOf(ev, dead) != player {
			continue
		}
		n++
	}
	return n
}

// gapB44MagmarchCandidates are the nonland permanents "another opponent"
// controls that a copy of `spell` (controlled by `caster`) could target in
// place of `from`, the permanent the spell targets: not controlled by the
// caster, not controlled by the controller of `from`, not a land, and a
// legal target of the spell's clause as it stands
// (game.SpellCouldTargetInsteadForEffect: hexproof, protection and a target
// clause that excludes it rule one out).
func gapB44MagmarchCandidates(g *game.Game, spell, caster, from uuid.UUID) []uuid.UUID {
	orig, ok := g.LookupCardForEffect(from)
	if !ok {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == from || c.IsLand() || c.Controller == caster || c.Controller == orig.Controller {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, from, c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// gapB44MagmarchSoleTarget is the one nonland permanent an opponent of
// `caster` controls that `spell` targets and nothing else: "targets only a
// single nonland permanent an opponent controls".
func gapB44MagmarchSoleTarget(g *game.Game, spell, caster uuid.UUID) (uuid.UUID, bool) {
	sole, ok := g.StackItemSoleTargetForEffect(spell)
	if !ok || sole.Kind != game.TargetCard {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(sole.ID)
	if !ok || c.IsLand() || !c.IsPermanent() || c.Controller == caster || !g.Battlefield.Contains(sole.ID) {
		return uuid.Nil, false
	}
	return sole.ID, true
}

// untapThisPermanentEffect is "Untap this creature (artifact)" as an
// activated ability's effect (Sage of the Maze, Honor-Worn Shaku): the source
// is untapped only while it is still the permanent that was activated
// (CR 400.7).
func untapThisPermanentEffect(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
