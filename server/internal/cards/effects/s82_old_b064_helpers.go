package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b064_helpers.go — shared pieces of Kaldheim batch 064 (#5505, S82).

// S82OldB064SnowPermanentsYouControl is how many snow permanents `player`
// controls, through the layered view ("if you control three or more snow
// permanents", Frost Bite and Blizzard Brawl).
func S82OldB064SnowPermanentsYouControl(g *game.Game, player uuid.UUID) int {
	return len(PermanentsControlledByMatching(g, player, Snow()))
}

// S82OldB064GreatestCreaturesSharingAType is Basalt Ravager's X: "the
// greatest number of creatures you control that have a creature type in
// common". Every creature type any of the creatures has is counted over the
// whole set (a creature with several types counts once for each), and a
// changeling, which has every type, counts for each of them; with only
// changelings the answer is their number. Zero with no creature.
func S82OldB064GreatestCreaturesSharingAType(g *game.Game, player uuid.UUID) int {
	var mine []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			mine = append(mine, c)
		}
	}
	types := map[string]bool{}
	changelings := 0
	for i := range mine {
		if game.HasAllCreatureTypes(&mine[i]) {
			changelings++
			continue
		}
		for _, t := range game.CreatureTypesOf(&mine[i]) {
			types[t] = true
		}
	}
	best := changelings
	for t := range types {
		n := 0
		for i := range mine {
			if mine[i].HasSubtype(t) {
				n++
			}
		}
		best = max(best, n)
	}
	return best
}

// S82OldB064OpponentsSnowLandsEnterTapped is Reidane's "Snow lands your
// opponents control enter tapped": the CR 614 replacement on the entering
// land's move, applied when its controller is not Reidane's.
func S82OldB064OpponentsSnowLandsEnterTapped(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller == src.Controller {
				return false
			}
			return entering.IsLand() && entering.HasSupertype("snow")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// S82OldB064PreventOneFromOpponents is Valkmira's prevention: "If a source an
// opponent controls would deal damage to you or a permanent you control,
// prevent 1 of that damage." The source is read as the damage event snapshotted
// it (CR 608.2h); a recipient is you, or a permanent you control on the
// battlefield.
func S82OldB064PreventOneFromOpponents() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			if !ok || ch.Controller == src.Controller || ch.Controller == uuid.Nil {
				return false
			}
			if ev.DamageTarget == src.Controller {
				return true
			}
			c, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && g.Battlefield.Contains(ev.DamageTarget) && c.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount--
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      "Valkmira, Protector's Shield — prevent 1 of that damage",
		RowLabel:   "If a source an opponent controls would deal damage to you or a permanent you control, prevent 1 of that damage.",
	}
}

// S82OldB064ValkmiraWard is Valkmira's second ability: "Whenever you or another
// permanent you control becomes the target of a spell or ability an opponent
// controls, counter that spell or ability unless its controller pays {1}."
// Ward's pay-or-counter body, over a player or any other permanent of yours.
func S82OldB064ValkmiraWard(label string) game.TriggeredAbility {
	cost := WardMana("{1}")
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.StackItemID == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
				return false
			}
			if ev.CardID == uuid.Nil {
				return ev.Target == source.Controller
			}
			if ev.CardID == source.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && g.Battlefield.Contains(ev.CardID) && c.Controller == source.Controller
		},
		Effect: WardEffect(cost),
	}
}

// --- Alrund, God of the Cosmos // Hakka, Whispering Raven ------------------

// S82OldB064AlrundCardTypes are the card types a player can name, in the order a type
// line lists them.
var S82OldB064AlrundCardTypes = []string{"Artifact", "Battle", "Creature", "Enchantment", "Instant", "Kindred", "Land", "Planeswalker", "Sorcery"}

// S82OldB064AlrundTypeChosen reveals the top two and takes every card of the chosen type,
// the rest going to the bottom in any order. An on-disk key: never renamed,
// never reused.
var S82OldB064AlrundTypeChosen = game.RegisterOptionPickThen("option-pick/s82-old-b064-alrund-card-type", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	want := r.Option.Label
	revealed := g.RevealTopOfLibraryForEffect(r.Chooser, r.Source, 2, "Alrund, God of the Cosmos — reveal the top two cards of your library")
	return TakeFromLibraryToHand{
		Player: r.Chooser,
		Cards:  revealed,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasCardType(strings.ToLower(want)) },
		All:    true,
		Label:  "Alrund, God of the Cosmos — " + want + " cards to your hand, the rest on the bottom",
		Key:    TakeRestBottomAny,
	}.applyKeyed(NewContext(g, nil), r.Chooser, r.Source)
})

// S82OldB064HakkaScry is "then scry 2" after the bounce. An on-disk key.
var S82OldB064HakkaScry = CardTail("then/s82-old-b064-hakka-scry", func(ctx *Context, _ game.TailOutcome) error {
	return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
