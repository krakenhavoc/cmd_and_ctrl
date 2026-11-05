package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S.S. Botany Bay — Legendary Artifact — Spacecraft {3}, 3/5 (Star Trek
// Commander, #4974):
//
//	"{T}: Each opponent exiles a card from their graveyard. You may
//	 exile a card from your graveyard.
//	 Station
//	 8+ | Flying
//	 Whenever S.S. Botany Bay deals combat damage to a player, return
//	 each creature card you own exiled with it to the battlefield."
//
// The {T} ability's exiles are LINKED to this object (CR 607.2a,
// ExileLinkedKeyedForEffect): each opponent picks a card of their own
// graveyard (Augusta, Order Returned's prompt run, APNAP order), those are
// exiled together, and then you may pick one of yours, which is exiled
// the same way. Station is the ordinary keyword; at eight charge counters
// it is a 3/5 artifact creature with flying and the combat-damage trigger
// (CR 721.2a, TriggeredAbility.ActiveWhen).
//
// The trigger reads the cards exiled with this object only, so a card
// exiled by another effect, or by an earlier Botany Bay object (it left
// and came back, CR 400.7), is not returned. "Creature card you own" is
// read in exile: an opponent's card it exiled never qualifies. They
// return together, under their owner's control.
//
// No simplification.
func init() {
	trigger := WheneverThisDealsCombatDamageToAPlayer("S.S. Botany Bay — return each creature card you own exiled with it to the battlefield", ssBotanyBayReturn)
	trigger.ActiveWhen = AtChargeCounters(8)
	Register(Spec{
		OracleID:     "8284c6ae-d86e-4f72-81da-7d3f9861fd12",
		Name:         "S.S. Botany Bay",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Each opponent exiles a card from their graveyard. You may exile a card from your graveyard.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  ssBotanyBayOpponentsExile,
			},
			Station(),
		},
		Static: append(
			SpacecraftAt(8, 3, 5),
			ThresholdKeywords(8, "flying"),
		),
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// ssBotanyBayOpponentsExile asks each opponent for a card from their own
// graveyard.
func ssBotanyBayOpponentsExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opponents := apnapOrder(g, ctx.Opponents())
	var legs []game.ChooseCardsPrompt
	for _, id := range opponents {
		cards := graveyardCardsOfPlayer(g, id)
		if len(cards) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "S.S. Botany Bay — exile a card from your graveyard",
			Cards:      cards,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneGraveyard,
		})
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, ssBotanyBayOpponentsChose, ctx.Source(), ctx.RunData(opponents))
	return err
}

// ssBotanyBayOpponentsChose exiles every opponent's pick at once, linked
// to this object, then asks you. An on-disk key.
var ssBotanyBayOpponentsChose = RunThen("prompt-run/ss-botany-bay-opponents-exile", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	var chosen []uuid.UUID
	for _, player := range r.Carry {
		chosen = append(chosen, picks.By(player)...)
	}
	link, ok := ctx.SourceRef()
	if len(chosen) == 0 || !ok {
		return ssBotanyBayAskYou(ctx)
	}
	return ctx.Game.ExileLinkedKeyedForEffect(link, chosen, ctx.ItemTail(ssBotanyBayThenYou))
})

// ssBotanyBayThenYou runs once the opponents' cards are exiled. An
// on-disk key.
var ssBotanyBayThenYou = CardTail("ss-botany-bay/then-you", func(ctx *Context, _ game.TailOutcome) error {
	return ssBotanyBayAskYou(ctx)
})

// ssBotanyBayAskYou is "You may exile a card from your graveyard."
func ssBotanyBayAskYou(ctx *Context) error {
	me := ctx.Controller()
	cards := graveyardCardsOfPlayer(ctx.Game, me)
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   "S.S. Botany Bay — you may exile a card from your graveyard",
		Cards:      cards,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    ssBotanyBayYouChose,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// ssBotanyBayYouChose exiles your pick, linked to this object. An on-disk
// key.
var ssBotanyBayYouChose = askedPick("choose-cards/ss-botany-bay-you-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0], Linked: true}.Apply(ctx)
})

// ssBotanyBayReturn returns each creature card its controller owns that
// is in exile linked to this object.
func ssBotanyBayReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var entries []game.BatchEntry
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.Owner == item.Controller && c.IsCreature() {
			entries = append(entries, game.BatchEntry{CardID: c.InstanceID, From: game.ZoneExile})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: item.Controller}, game.KeyedTail{})
}
