package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soulbond.go — the soulbond keyword (CR 702.95, #5675) as a card file
// spells it:
//
//	PrintedKeywords: []string{game.KeywordSoulbond},
//	Triggered:       Soulbond("Doom Weaver"),
//	Static:          []game.StaticAbility{PairedCreaturesHave(label, grantKey)},
//
// Soulbond is two triggered abilities (CR 702.95a):
//
//	"When this creature enters, if you control both this creature and
//	 another creature and both are unpaired, you may pair this creature
//	 with another unpaired creature you control for as long as both
//	 remain creatures on the battlefield under your control"
//	"Whenever another creature you control enters, if you control both
//	 that creature and this one and both are unpaired, you may pair
//	 that creature with this creature for as long as both remain
//	 creatures on the battlefield under your control."
//
// Each has an intervening "if", so it is judged as the ability would
// trigger and again as it resolves (CR 603.4): a creature another
// soulbond ability paired in the meantime is not offered, and an
// ability with nothing left to pair asks nothing. The partner is not a
// target. It is chosen as the ability resolves, on the "choose
// permanents" prompt with a floor of zero, which is the "you may".
//
// The rows read the soulbond token off the layered ability list, so a
// creature that has lost soulbond (a token copy of Mirage Phalanx,
// "except it … loses soulbond") never asks. The pair itself, and what
// ends it (CR 702.95e), is game/soulbond.go.

// Soulbond is the keyword's two triggered abilities. The card also
// declares game.KeywordSoulbond in its PrintedKeywords.
func Soulbond(cardName string) []game.TriggeredAbility {
	return []game.TriggeredAbility{
		On(game.EventETB, soulbondThisEntered, cardName+" — soulbond: pair it with another unpaired creature you control", soulbondPairThis),
		On(game.EventETB, soulbondAnotherEntered, cardName+" — soulbond: pair it with the creature that entered", soulbondPairEntered),
	}
}

// soulbondThisEntered is the first ability's condition: this creature
// entered, it has soulbond, and its controller controls another
// unpaired creature beside it.
func soulbondThisEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID != source.InstanceID || !game.HasKeyword(source, game.KeywordSoulbond) {
		return false
	}
	return len(soulbondPartners(g, source.Controller, source.InstanceID)) > 0
}

// soulbondAnotherEntered is the second ability's condition: another
// creature entered under this creature's controller's control, this
// creature has soulbond, and the two may be paired.
func soulbondAnotherEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsCreature() || !game.HasKeyword(source, game.KeywordSoulbond) {
		return false
	}
	return g.CanPairForEffect(source.Controller, source.InstanceID, c.InstanceID)
}

// soulbondPartners lists what `self` may be paired with now: nothing
// when self is not an unpaired creature `controller` controls,
// otherwise that player's other unpaired creatures.
func soulbondPartners(g *game.Game, controller, self uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range g.UnpairedCreaturesForEffect(controller, self) {
		if g.CanPairForEffect(controller, self, id) {
			out = append(out, id)
		}
	}
	return out
}

// soulbondPairThis resolves the first ability: the controller may pair
// the soulbond creature with one of their other unpaired creatures.
func soulbondPairThis(g *game.Game, item *game.StackItem) error {
	return soulbondAsk(g, item, "Soulbond — pair "+soulbondName(g, item.SourceCardID)+" with another unpaired creature you control?",
		soulbondPartners(g, item.Controller, item.SourceCardID))
}

// soulbondPairEntered resolves the second ability: the controller may
// pair the creature that entered with the soulbond creature.
func soulbondPairEntered(g *game.Game, item *game.StackItem) error {
	entered := item.Trigger.Event.CardID
	if !g.CanPairForEffect(item.Controller, item.SourceCardID, entered) {
		return nil
	}
	return soulbondAsk(g, item, "Soulbond — pair "+soulbondName(g, entered)+" with "+soulbondName(g, item.SourceCardID)+"?",
		[]uuid.UUID{entered})
}

// soulbondName is a permanent's name for the question.
func soulbondName(g *game.Game, id uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(id); ok && c.Name != "" {
		return c.Name
	}
	return "this creature"
}

// soulbondAsk queues the "you may pair" question over `partners`, or
// does nothing when there are none (CR 603.4).
func soulbondAsk(g *game.Game, item *game.StackItem, question string, partners []uuid.UUID) error {
	if len(partners) == 0 {
		return nil
	}
	return ChoosePermanents{
		Question: question,
		Candidates: func(*game.Game, uuid.UUID) ([]uuid.UUID, int, int) {
			return partners, 0, 1
		},
		Keyed: soulbondPicked,
		Carry: []uuid.UUID{item.SourceCardID},
	}.Apply(NewContext(g, item))
}

// soulbondPicked is the rest of a soulbond ability: the soulbond
// creature (the run's carry) and the creature chosen become paired, if
// both are still unpaired creatures the ability's controller controls
// (CR 702.95c). An on-disk key.
var soulbondPicked = RegisterPermanentsPick("soulbond/pair", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 || len(r.Carry) == 0 {
			return nil
		}
		ctx.Game.PairForEffect(ctx.Controller(), r.Carry[0], cards[0])
		return nil
	},
})

// PairedCreaturesHave is "As long as <this> is paired with another
// creature, each of those creatures has '<ability>'": a layer-6 grant
// (ADR 0093) of the named bundles to the source and to the creature it
// is paired with, for as long as the pair lasts. `label` is the row's
// wording in the static-ability chip.
func PairedCreaturesHave(label string, keys ...string) game.StaticAbility {
	ab := GrantAbilities(pairedWithSource, keys...)
	ab.Label = label
	return ab
}

// pairedWithSource is "each of those creatures": the source while it is
// paired, and its partner.
func pairedWithSource(target *game.Card, g *game.Game, source *game.Card) bool {
	if source == nil {
		return false
	}
	p := g.PairedPartner(source)
	return p != nil && (target.InstanceID == source.InstanceID || target.InstanceID == p.InstanceID)
}
