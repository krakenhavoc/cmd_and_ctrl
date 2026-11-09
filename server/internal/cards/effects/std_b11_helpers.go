package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b11_helpers.go — shared vocabulary of the Standard leftovers,
// batches 11 and 12 (#4452).

// sb11ColorBits gives each colour a bit of a SearchRule's number.
var sb11ColorBits = map[string]int{"W": 1, "U": 2, "B": 4, "R": 8, "G": 16}

// sb11ColorMask is the number of searchRuleOneCardPerColor naming these
// colours.
func sb11ColorMask(colors ...string) int {
	m := 0
	for _, c := range colors {
		m |= sb11ColorBits[c]
	}
	return m
}

// searchRuleOneCardPerColor is "search your library for a black card, a
// green card, and a blue card" (Lotuslight Dancers): the number names the
// colours, and the picked set must give every card its own colour, a card
// standing for one quality only. A card of several colours may be any one
// of them, and the others still need cards of their own (Lotuslight
// Dancers' ruling). Fewer cards than colours is fine, none at all too.
var searchRuleOneCardPerColor = game.RegisterSearchRule("search/one-card-per-color",
	func(picked []game.Card, mask int) bool {
		var colors []string
		for c, bit := range sb11ColorBits {
			if mask&bit != 0 {
				colors = append(colors, c)
			}
		}
		if len(picked) > len(colors) {
			return false
		}
		return sb11Assign(picked, colors, map[string]bool{})
	})

// sb11Assign reports whether every card in picked can take a distinct
// colour from colors.
func sb11Assign(picked []game.Card, colors []string, used map[string]bool) bool {
	if len(picked) == 0 {
		return true
	}
	for _, col := range colors {
		if used[col] || !picked[0].HasColor(col) {
			continue
		}
		used[col] = true
		ok := sb11Assign(picked[1:], colors, used)
		used[col] = false
		if ok {
			return true
		}
	}
	return false
}

// --- Call the Spirit Dragons: one +1/+1 counter per colour -----------------

// sb11CallColors is the order Call the Spirit Dragons walks the colours.
var sb11CallColors = []string{"W", "U", "B", "R", "G"}

// sb11CallPick is one leg of the chain: the Dragon chosen for the colour
// whose index rides CarryNums, the Dragons counted so far Carry.
var sb11CallPick game.ChooseCardsThen

// Registered from init, not at the var: the leg and sb11CallStep name each
// other.
func init() {
	sb11CallPick = game.RegisterChooseCardsThen("choose-cards/call-the-spirit-dragons-pick", sb11CallPickLeg)
}

func sb11CallPickLeg(g *game.Game, r game.CardsChosen) error {
	if len(r.Nums) != 1 {
		return nil
	}
	given := append([]uuid.UUID(nil), r.Carry...)
	if len(r.Picked) == 1 {
		var err error
		if given, err = sb11CallPut(g, r.Picked[0], given); err != nil {
			return err
		}
	}
	return sb11CallStep(g, r.Chooser, r.Source, r.Word, r.Nums[0]+1, given)
}

// sb11CallPut puts the +1/+1 counter on the Dragon and counts it once.
func sb11CallPut(g *game.Game, id uuid.UUID, given []uuid.UUID) ([]uuid.UUID, error) {
	if _, ok := battlefieldCardForEffect(g, id); !ok {
		return given, nil
	}
	if err := g.AddCounterForEffect(id, game.CounterPlusOne, 1); err != nil {
		return given, err
	}
	for _, d := range given {
		if d == id {
			return given, nil
		}
	}
	return append(given, id), nil
}

// sb11CallStep carries on from colour index `from`, `dragon` being the
// Dragon type word as the enchantment's text has it (CR 612.2): a colour with one
// Dragon of it takes its counter at once, one with several asks the
// controller, one with none is skipped. After the last colour the
// controller wins if five different Dragons got counters.
func sb11CallStep(g *game.Game, controller, source uuid.UUID, dragon string, from int, given []uuid.UUID) error {
	for i := from; i < len(sb11CallColors); i++ {
		var cands []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.IsCreature() && c.HasSubtype(dragon) && c.HasColor(sb11CallColors[i]) {
				cands = append(cands, c.InstanceID)
			}
		}
		switch len(cands) {
		case 0:
		case 1:
			var err error
			if given, err = sb11CallPut(g, cands[0], given); err != nil {
				return err
			}
		default:
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:   controller,
				Source:    source,
				Question:  "Call the Spirit Dragons — put a +1/+1 counter on a " + sb11CallColors[i] + " Dragon you control",
				Cards:     cands,
				Min:       1,
				Max:       1,
				ThenKey:   sb11CallPick,
				Carry:     given,
				CarryNums: []int{i},
				CarryWord: dragon,
				LiveZone:  game.ZoneBattlefield,
			})
			return nil
		}
	}
	if len(given) >= 5 {
		_, err := g.WinTheGameForEffect(controller, source)
		return err
	}
	return nil
}

// --- "You may behold a <type>" as an instruction ------------------------------

// MayBehold is "you may behold <article> <subtype>. If you do, ..." as an
// instruction of a resolving ability (CR 701.4a), where OptionalBehold is
// the additional cost of a spell. To behold a creature type is to choose a
// permanent of that type you control or reveal a card of that type from
// your hand: one choose_cards prompt over both, 0 or 1 cards. A card from
// the hand is revealed to the table; a permanent is only chosen. Whatever
// the pick, "if you do" runs once it was made (CR 701.4b: a fact about the
// choice, not about the card afterwards). With nothing to behold the
// player is not asked and nothing happens.
//
// Build it once as a package-level var in the card's file; its key is an
// on-disk identity, "choose-cards/<card>-behold", never renamed or reused.
type MayBehold struct {
	name, subtype string
	pick          game.ChooseCardsThen
}

// NewMayBehold registers the keyed prompt. ifYouDo runs against the asking
// item's context when the controller beholds.
func NewMayBehold(name, subtype, key string, ifYouDo func(ctx *Context) error) *MayBehold {
	b := &MayBehold{name: name, subtype: subtype}
	b.pick = askedPick(key, func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		g := ctx.Game
		if p := g.PlayerByIDForEffect(r.Chooser); p != nil && p.Hand != nil && p.Hand.Contains(r.Picked[0]) {
			g.RevealForEffect(game.RevealSpec{
				Player: r.Chooser, Source: ctx.Source(),
				Reason: name + " — behold a " + subtype, Cards: r.Picked,
			})
		}
		return ifYouDo(ctx)
	})
	return b
}

// Ask queues the offer.
func (b *MayBehold) Ask(ctx *Context) error {
	g := ctx.Game
	me := ctx.Controller()
	var cands []uuid.UUID
	word := ctx.TextSubtype(b.subtype)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me && c.HasSubtype(word) {
			cands = append(cands, c.InstanceID)
		}
	}
	if p := g.PlayerByIDForEffect(me); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if c.HasSubtype(word) {
				cands = append(cands, c.InstanceID)
			}
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: b.name + " — you may behold a " + b.subtype + " (choose one you control or reveal one from your hand)",
		Cards:    cands,
		Min:      0,
		Max:      1,
		ThenKey:  b.pick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}
