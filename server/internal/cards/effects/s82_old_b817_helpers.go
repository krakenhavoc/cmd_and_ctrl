package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b817_helpers.go — Alliances, instants and sorceries (batch 817).

// OldB817PlainsOrWhitePermanent is "a Plains or a white permanent".
func OldB817PlainsOrWhitePermanent() CardPredicate {
	return Or(HasSubtype("Plains"), OfColor("W"))
}

// OldB817EachPlayerSacrificesForEachWhitePermanent is Omen of Fire's last
// sentence: "Each player sacrifices a Plains or a white permanent of their
// choice for each white permanent they control." Each living player, in
// APNAP order, counts the white permanents they control as the sentence
// begins and makes ONE choice of that many Plains or white permanents (every
// white permanent qualifies, so there are always enough); the choices are
// sacrificed together. A player with no white permanent is not asked. The
// choice is not a target.
func OldB817EachPlayerSacrificesForEachWhitePermanent(g *game.Game, item *game.StackItem) error {
	source := item.SourceCardID
	spec := SacrificeSpec("a Plains or a white permanent", OldB817PlainsOrWhitePermanent())
	white := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasColorWord("W") {
			white[c.Controller]++
		}
	}
	var legs []game.ChooseCardsPrompt
	for _, seat := range ApnapLivingSeats(g) {
		n := white[seat]
		if n == 0 || !g.CanEffectCauseSacrificeForEffect(seat) {
			continue
		}
		candidates := g.EffectSacrificeCandidatesForEffect(seat, spec)
		if len(candidates) == 0 {
			continue
		}
		k := min(n, len(candidates))
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser: seat, FromPlayer: seat, Source: source,
			Question: "Omen of Fire — sacrifice a Plains or a white permanent for each white permanent you control",
			Cards:    candidates, Min: k, Max: k, Zone: game.ZoneBattlefield,
			Destination: game.ChooseSacrifice,
		})
	}
	return g.ChoosePermanentsInOrderSacrificeForEffect(source, legs)
}

// OldB817RemoveCountersAtCleanupBody is Bounty of the Hunt's delayed
// trigger: "remove a +1/+1 counter from that creature" once for each
// counter the spell put on it. It names the creature as the OBJECT it was
// (CR 400.7) and takes the count from Params.Amount; a creature that has
// fewer counters than that loses the ones it has (CR 122.1).
var OldB817RemoveCountersAtCleanupBody = game.DelayedBody("bounty-of-the-hunt/remove-counters", oldB817RemoveCounters)

func oldB817RemoveCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(p.Object.ID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	n := min(p.Amount, c.Counters[game.CounterPlusOne])
	if n <= 0 {
		return nil
	}
	return g.AddCounterForEffect(p.Object.ID, game.CounterPlusOne, -n)
}
