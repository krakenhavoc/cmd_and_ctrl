package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b115_helpers.go — shared pieces of Ikoria: Lair of Behemoths
// batch 115 (#5512).

// S82OldB115AnotherSpellWithFlashCastByYou is "whenever you cast another spell
// that has flash" (Slitherwisp): the spell is read where it sits on the
// stack, with the keyword as it was cast.
func S82OldB115AnotherSpellWithFlashCastByYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventCast || ev.Actor != source.Controller || ev.CardID == source.InstanceID {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.HasKeyword(&spell, "flash")
}

// S82OldB115CreatureSpellWithMutateCastByYou is "whenever you cast a creature
// spell, if it has mutate" (Pollywog Symbiote).
func S82OldB115CreatureSpellWithMutateCastByYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsCreature() && game.HasKeyword(&spell, "mutate")
}

// S82OldB115MonstrousStep is Monstrous Step's resolution: the first target
// gets +7/+7 until end of turn, and the second, if there is one and it is
// still legal, blocks the first this turn if able.
func S82OldB115MonstrousStep(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	first := item.Targets[0]
	if !ctx.IsTargetLegal(first) {
		return nil
	}
	if err := (BoostUntilEOT{Target: first.ID, Power: 7, Toughness: 7, Label: "Monstrous Step — +7/+7"}).Apply(ctx); err != nil {
		return err
	}
	if len(item.Targets) < 2 || !ctx.IsTargetLegal(item.Targets[1]) {
		return nil
	}
	ref, ok := ctx.Game.PermanentRefForEffect(first.ID)
	if !ok {
		return nil
	}
	return BlocksAttackerUntilEOT{
		Blocker:  item.Targets[1].ID,
		Attacker: ref,
		Label:    "Monstrous Step — blocks the other target creature if able",
	}.Apply(ctx)
}

// S82OldB115InstantsAndSorceriesInYourGraveyard counts the instant and
// sorcery cards in the source's controller's graveyard (Rielle).
func S82OldB115InstantsAndSorceriesInYourGraveyard(g *game.Game, source *game.Card) int {
	p := g.PlayerByIDForEffect(source.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.HasCardType("instant") || c.HasCardType("sorcery") {
			n++
		}
	}
	return n
}

// S82OldB115DiscardedForTheFirstTimeThisTurn is "whenever you discard one or
// more cards for the first time each turn" (Rielle): a discard by you with no
// discard by you in an earlier batch of the turn. Every discard of the first
// batch is accepted, so the batch's count is the number of cards drawn.
func S82OldB115DiscardedForTheFirstTimeThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !DiscardedByYou(ev, source) {
		return false
	}
	for _, e := range g.EventsThisTurn() {
		if e.Kind == game.EventDiscardCard && e.Actor == source.Controller && e.Batch < ev.Batch {
			return false
		}
	}
	return true
}
