package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shared helpers for Avatar batches 26 to 29 (#5075).

// avatar_b26_part27_helpers.go — shared pieces of Avatar: The Last
// Airbender Eternal batch 27 (#5075).

// avB27SwampsOnTheBattlefield is "the number of Swamps on the
// battlefield" (Swampbenders): every land with the subtype under any
// controller, read through the layered type line. The caller reads the
// subtype through its own text (source.TextSubtype).
func avB27SwampsOnTheBattlefield(g *game.Game, swamp string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.HasSubtype(swamp) {
			n++
		}
	}
	return n
}

// avB27LandsYouControlAreAlso is "Lands you control are <subtype>s in
// addition to their other types" (Swampbenders): EachLandIsAlso narrowed
// to the source's controller's lands. Layer 4, an idempotent append.
func avB27LandsYouControlAreAlso(subtype string) game.StaticAbility {
	ab := EachLandIsAlso(subtype)
	ab.Label = "Lands you control are " + subtype + "s in addition to their other types."
	ab.AppliesTo = func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsLand() && target.Controller == source.Controller
	}
	return ab
}

// avB27AlliesYouControlHaveDuringYourTurn is "During your turn, Allies
// you control have <keyword>" (Sokka's Charge): a layer-6 grant over the
// Ally creatures the source's controller controls while that player is
// the active player.
func avB27AlliesYouControlHaveDuringYourTurn(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() &&
			target.HasSubtype(source.TextSubtype("Ally")) && isActivePlayer(g, source.Controller)
	}, keyword)
	ab.Label = "During your turn, Allies you control have " + keyword + "."
	return ab
}

// avB27TappedArtifactOrCreature is "tapped artifacts and/or creatures"
// (Aang and Katara): a tapped permanent that is an artifact, a creature
// or both, counted once.
func avB27TappedArtifactOrCreature(c game.Card) bool {
	return c.Tapped && (c.IsArtifact() || c.IsCreature())
}

// avB27CondThisDidntAttack is "if this creature didn't attack this
// turn" (Air Nomad Student) as an intervening-if condition, re-checked
// as the item resolves (CR 603.4). It declares FootprintCondAttacked:
// the attack record only changes by declaring attackers, which no item
// on the stack does.
type avB27CondThisDidntAttack struct{}

func (avB27CondThisDidntAttack) Holds(ctx *Context) bool {
	return !ctx.Game.AttackedThisTurn(ctx.Source())
}

func (avB27CondThisDidntAttack) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondAttacked, 0), true
}

// avB27LandCreaturesYouControlHave is "Land creatures you control have
// <keyword>" (Toph, Greatest Earthbender): a layer-6 grant over each
// permanent the source's controller controls that is both a land and a
// creature, re-read every layer pass.
func avB27LandCreaturesYouControlHave(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsLand() && target.IsCreature()
	}, keyword)
	ab.Label = "Land creatures you control have " + keyword + "."
	return ab
}

// avatar_b26_part28_helpers.go — shared pieces for Avatar: The Last
// Airbender Eternal (#5075), batch 26 (part 28).

// avB28CreatureLeftUnderYourControl is "a creature left the battlefield
// under your control this turn" (Tale of Momo, That's Rough Buddy): the
// per-player CreaturesLeft tally, so a creature that left by any route,
// to any zone, earlier in the turn counts, tokens included.
func avB28CreatureLeftUnderYourControl(g *game.Game, you uuid.UUID) bool {
	return g != nil && g.CreatureLeftThisTurn(you)
}

// avB28DrawThenDiscardUnlessThreeLessons is Chakra Meditation's "Then
// discard a card unless there are three or more Lesson cards in your
// graveyard": the graveyard is counted after the draw, as the discard
// would be made. The key is an on-disk identity.
var avB28DrawThenDiscardUnlessThreeLessons = game.RegisterDrawThen("discard-unless-three-lessons", func(g *game.Game, d game.DrawThen) error {
	if avB4LessonsInGraveyard(g, d.Player) >= 3 {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
	return nil
})

// avatar_b26_part29_helpers.go — shared pieces of Avatar: The Last
// Airbender Eternal batch 29 (#5075).

// avB29PutIntoALibraryEvents are the events a card being put into a
// library arrives as: an ordinary zone move (a tuck, a shuffle-in, a
// card from outside the game), a discard a replacement sent to the
// library (Library of Leng: still EventDiscardCard, NewZone the library),
// and a counterspell that put the spell into its owner's library
// (Memory Lapse: EventCounterSpell, which names no destination).
var avB29PutIntoALibraryEvents = []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventCounterSpell}

// avB29ACardWasPutIntoALibrary is "a card is put into a library from
// anywhere" (Dutiful Knowledge Seeker, Wan Shi Tong, All-Knowing): the
// card arrived in a library from another zone. A move inside a library
// (scry, surveil's top half, any reordering) is not one (the ruling), and
// a token is not a card (CR 108.2): one that is put into a library is
// there until the next state check removes it (CR 704.5d), so it is
// looked up and refused. Read as the event is emitted, so the card is
// where the move put it.
func avB29ACardWasPutIntoALibrary(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventZoneMove, game.EventDiscardCard:
		if ev.NewZone != game.ZoneLibrary || ev.OldZone == game.ZoneLibrary {
			return false
		}
	case game.EventCounterSpell:
	default:
		return false
	}
	z := g.FindCardZoneForEffect(ev.CardID)
	if z == nil || z.Kind != game.ZoneLibrary {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// avB29WheneverCardsArePutIntoALibrary is "Whenever one or more cards
// are put into a library from anywhere, <effect>": ONE trigger for every
// card the same event batch put into libraries (CR 603.2c), however many
// cards and libraries — the batch-settled shape of "whenever one or more
// nonland cards are milled" (nonland_cards_milled.go).
func avB29WheneverCardsArePutIntoALibrary(label string, effect Effect) game.TriggeredAbility {
	return withAtBatchEnd(OnAny(avB29PutIntoALibraryEvents, avB29ACardWasPutIntoALibrary, label, effect), ThatManyEvents())
}

// avB29NightmaresMillBody is Nightmares and Daydreams' delayed trigger:
// "target player mills cards equal to that spell's mana value". The
// spell rides as the item's payload; its mana value is read as the
// trigger resolves, with its announced X while it is still on the stack
// (triggeringSpellManaValue's reading). An on-disk key.
var avB29NightmaresMillBody = game.TargetedDelayedBody("nightmares-and-daydreams/target-player-mills",
	simpleBody(avB29NightmaresMill),
	constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))

func avB29NightmaresMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	spell := firstPayloadCard(ctx)
	if victim == uuid.Nil || spell == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(spell)
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(c)
	return MillCards{Player: victim, N: mv}.Apply(ctx)
}

// avB29NightmaresChapter is chapters I, II and III: "Until your next
// turn, whenever you cast an instant or sorcery spell, target player
// mills cards equal to that spell's mana value." A delayed trigger
// (CR 603.7) that fires on every such cast (Repeats) until the
// controller's next turn begins, held by the chapter's controller
// (CR 603.7d). It is data — a registered condition and body — so a table
// with it waiting is still a restore point.
func avB29NightmaresChapter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	until := DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Nightmares and Daydreams — target player mills cards equal to that spell's mana value",
		On:           []game.EventKind{game.EventCast},
		Condition:    youNextCastCondition,
		CondParams:   game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}},
		Body:         avB29NightmaresMillBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// avB29AGraveyardHasTwenty is "a graveyard has twenty or more cards in
// it": any player's.
func avB29AGraveyardHasTwenty(g *game.Game) bool {
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 20 {
			return true
		}
	}
	return false
}
