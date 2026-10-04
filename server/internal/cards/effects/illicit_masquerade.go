package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illicit Masquerade — Enchantment {3}{B}:
//
//	"Flash
//	 When this enchantment enters, put an impostor counter on each
//	 creature you control.
//	 Whenever a creature you control with an impostor counter on it
//	 dies, exile it. Return up to one other target creature card from
//	 your graveyard to the battlefield."
//
// The counters are read off the departure record (CR 603.10a): MoveCard
// clears them on the way out, so "with an impostor counter on it" is
// the permanent as it last existed, the same reading Toxrill's slime
// counters get. Controller is the one the creature had as it died.
//
// "Exile it" is the card that died, and only while it is still that
// object in a graveyard (CR 400.7): one reanimated or moved in
// response is a new object and is left alone. The return happens
// either way. "Other" is that card, not this enchantment, so the clause
// is built per trigger (TargetsFrom) from the triggering event; "up to
// one" means the trigger still goes on the stack with an empty
// graveyard (CR 603.3d does not remove it).
//
// No simplification.
const illicitMasqueradeCounter = "impostor"

func init() {
	Register(Spec{
		OracleID:        "fa2a275a-a28b-45cd-b8ca-a93247ff59bc",
		Name:            "Illicit Masquerade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Illicit Masquerade — put an impostor counter on each creature you control",
				putACounterOnEachCreatureYouControl(illicitMasqueradeCounter)),
			{
				Watches:     []game.EventKind{game.EventLTB},
				AppliesTo:   illicitMasqueradeImpostorDied,
				TargetsFrom: illicitMasqueradeOtherCreatureCard,
				Key:         "Illicit Masquerade — exile it, then return up to one other creature card from your graveyard to the battlefield",
				Effect:      illicitMasqueradeExileAndReturn,
			},
		},
	})
}

// illicitMasqueradeImpostorDied: a creature you controlled died with an
// impostor counter on it.
func illicitMasqueradeImpostorDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(ev.CardID)
	return ok && info.Counters[illicitMasqueradeCounter] > 0
}

// illicitMasqueradeOtherCreatureCard is "up to one other target
// creature card from your graveyard", where "other" is the card that
// died.
func illicitMasqueradeOtherCreatureCard(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return TargetCardInGraveyard("up to one other target creature card from your graveyard",
		Creature(), YouOwn(), NotSelf(tc.Event.CardID)).WithCount(0, 1)
}

func illicitMasqueradeExileAndReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if dead := item.Trigger.Event.CardID; diedCardStillInGraveyard(ctx, dead) {
		if err := (ExileTarget{Target: dead}).Apply(ctx); err != nil {
			return err
		}
	}
	return returnFirstLegalGraveyardTargetToBattlefield(g, item)
}
