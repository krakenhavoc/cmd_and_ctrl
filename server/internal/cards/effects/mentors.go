package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// mentors.go — "mentors a creature" (#5932, CR 702.134c). The keyword's
// trigger is Mentor (source_relative_targets.go); its body is
// mentorResolves, which records game.EventMentored for the creature it
// targeted before putting the counter on it, so another permanent can
// watch: Aegis of the Legion's "Whenever equipped creature mentors a
// creature, put a shield counter on that creature."

// mentorResolves is the mentor ability's resolution: the target, if it is
// still legal, was mentored by the ability's source (CR 702.134c), and
// gets a +1/+1 counter. The event comes first: the ability resolved
// whatever a replacement then does with the counter.
func mentorResolves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.MentoredForEffect(item.SourceCardID, item.Controller, id)
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// equippedCreatureMentored is "whenever equipped creature mentors a
// creature": the mentor ability's source is the creature this Equipment
// is attached to as the ability resolves.
func equippedCreatureMentored(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventMentored && source.IsAttachedTo(ev.Source)
}

// WheneverEquippedCreatureMentors is the printed "Whenever equipped
// creature mentors a creature, …" (Aegis of the Legion). The effect reads
// the mentored creature as ctx.TriggeringPermanent (the event's CardID).
func WheneverEquippedCreatureMentors(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventMentored, equippedCreatureMentored, label, effect)
}
