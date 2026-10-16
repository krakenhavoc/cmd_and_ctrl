package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rinoa, Angel Wing — Legendary Creature — Human Rebel Warlock {2}{W},
// 2/4 (the Squall deck request, #3918):
//
//	"At the beginning of combat on your turn, creatures you control with
//	 flying get +1/+1 and gain vigilance until end of turn.
//	 Whenever one or more attacking creatures you control die, you may
//	 return one of them to the battlefield tapped under its owner's
//	 control with a flying counter on it. Do this only once each turn."
//
// The combat trigger's set is fixed as it resolves (CR 611.2c): the
// creatures you control with flying then get both halves, and a creature
// that gains flying later in the turn does not.
//
// The death trigger is a batch trigger (CR 603.2c): one per event that
// kills one or more of your attackers, so a block that kills three is
// one trigger. "Attacking" and "you control" are read as the creatures
// last existed (CR 603.10a, diedWhileAttacking and leftUnderControlOf),
// so an attacker that dies to a removal spell after blockers counts and
// a stolen attacker counts for its thief. As it resolves, "them" is the
// batch's attackers still in a graveyard; you may pick one, and it
// returns tapped under its OWNER's control with a flying counter
// (CR 122.1b).
//
// "Do this only once each turn" limits the DOING, as Krile Baldesion's
// does: a trigger whose "may" is declined, or whose creatures have all
// left the graveyard, leaves the turn's one return for a later batch.
// The use is recorded only when a creature is returned, read when the
// ability triggers (no trigger once it is spent) and again as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47b29cf9-8ef3-4b6a-a9a3-3ab822c5dea4",
		Name:         "Rinoa, Angel Wing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Rinoa, Angel Wing — creatures you control with flying get +1/+1 and gain vigilance",
				rinoaFliersPump),
			OncePerBatch(On(game.EventLTB, rinoaAttackerYouControlDied, rinoaReturnLabel, rinoaReturnOne)),
		},
	})
}

const (
	rinoaReturnLabel = "Rinoa, Angel Wing — you may return one of them with a flying counter"
	// rinoaDoneLabel is the tally the once-each-turn limit reads.
	rinoaDoneLabel = "Rinoa, Angel Wing — return (done)"
)

func rinoaFliersPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	fliers := And(Creature(), YouControl(), HasKeyword("flying"))
	// The set is read once, before either half applies, so both halves
	// land on the same creatures.
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if fliers(g, item.Controller, c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Rinoa, Angel Wing — +1/+1"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Rinoa, Angel Wing — vigilance"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rinoaAttackerYouControlDied is one death in the batch: a creature that
// was attacking under Rinoa's controller as it died, while the turn's
// one return is unspent.
func rinoaAttackerYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if g.ResolvedThisTurn(source.InstanceID, rinoaDoneLabel) > 0 {
		return false
	}
	dead, ok := diedWhileAttacking(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller
}

// rinoaReturnOne offers the batch's attackers still in a graveyard and
// returns the one picked.
func rinoaReturnOne(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || g.ResolvedThisTurn(item.SourceCardID, rinoaDoneLabel) > 0 {
		return nil
	}
	batch := item.Trigger.Event.Batch
	source := game.Card{InstanceID: item.SourceCardID, Controller: item.Controller}
	var candidates []uuid.UUID
	owners := map[uuid.UUID]bool{}
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		// An unstamped event (Batch zero) is a batch of one: itself.
		sameBatch := ev.Batch == batch && (batch != 0 || ev.Seq == item.Trigger.Event.Seq)
		if !sameBatch || seen[ev.CardID] {
			continue
		}
		dead, ok := diedWhileAttacking(ev, g)
		if !ok || leftUnderControlOf(ev, dead) != source.Controller {
			continue
		}
		seen[ev.CardID] = true
		z := g.FindCardZoneForEffect(ev.CardID)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, ev.CardID)
		owners[z.Owner] = true
	}
	if len(candidates) == 0 {
		return nil
	}
	rinoa := item.SourceCardID
	prompt := game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   rinoa,
		Question: "Rinoa, Angel Wing — you may return one of them to the battlefield tapped with a flying counter",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) != 1 || g.ResolvedThisTurn(rinoa, rinoaDoneLabel) > 0 {
				return nil
			}
			_, err := g.ReturnFromGraveyardWithCountersForEffect(picked[0], uuid.Nil, true,
				map[string]int{game.CounterFlying: 1})
			if errors.Is(err, game.ErrCardNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			// The return happened when the card left the graveyard (an
			// entry that paused on a replacement prompt has left it too).
			if z := g.FindCardZoneForEffect(picked[0]); z == nil || z.Kind != game.ZoneGraveyard {
				g.NoteDoneThisTurn(rinoa, rinoaDoneLabel)
			}
			return nil
		},
	}
	// The live re-check on submit reads one owner's graveyard. Every
	// candidate is almost always in yours; when they span graveyards
	// (an attacker you stole), the return itself re-checks the zone.
	if len(owners) == 1 {
		for owner := range owners {
			prompt.FromPlayer = owner
		}
		prompt.Zone = game.ZoneGraveyard
	}
	g.QueueChooseCardsForEffect(prompt)
	return nil
}
