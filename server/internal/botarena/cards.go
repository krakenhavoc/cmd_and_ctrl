package botarena

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cards.go is the arena's Cards section (ADR 0126 §1 item 2): for each
// contestant, every non-land card its seats were offered, how often,
// and how often they took it.
//
// "Offered" means a cast or activate move naming the card was in the
// window's move list. "Taken" means the move the runner dispatched,
// and the engine applied, was one of them. Both are counted per
// window, and a window counts a card once however many ways it offers
// it (two targets, two modes, two abilities): the question is "was
// the card on the table to be used", not "how many shapes did the
// enumerator find".
//
// The tally comes from the runner's observer, the same feed the
// decision log writes, so it needs no log on disk. It reads only the
// seat's own filtered view and move list, which the runner hands the
// observer anyway.
//
// Only the seat's OWN cards are counted (a card it owns, not a token,
// not a stolen card it was let cast from exile), and only non-land
// ones: a land's cast is a land drop, which is KindLand and never
// here, and an activated ability on a land is not what ADR 0126 asks
// about.

// Card actions the Cards section tallies, from legal.Kind.
const (
	ActionCast     = "cast"
	ActionActivate = "activate"
)

// NeverWindows is how many windows a card must have been offered in,
// and never taken, to be flagged `never` (ADR 0126 §1).
const NeverWindows = 5

// CardUse is one seat's use of one card in one game.
type CardUse struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	// ManaSource is true when the card carries a REPEATABLE mana
	// ability, ADR 0126 §2's definition: a tap ability with no
	// sacrifice or exile cost. That is a rock or a dork, acceptance
	// bar A2's class. A Treasure-shaped one-shot (Lotus Petal) and a
	// sacrifice-another source (Ashnod's Altar) are not.
	ManaSource bool `json:"mana_source,omitempty"`
	// Offered is the windows the card was offered in; Taken is the
	// windows in which the seat used it.
	Offered int `json:"offered"`
	Taken   int `json:"taken"`
}

type cardKey struct {
	name, action string
}

// cardTally is one seat's observer for the Cards section.
type cardTally struct {
	mu   sync.Mutex
	uses map[cardKey]*CardUse
}

func newCardTally() *cardTally { return &cardTally{uses: map[cardKey]*CardUse{}} }

// Observe implements aiseat.DecisionObserver.
func (t *cardTally) Observe(ev aiseat.DecisionEvent) {
	var (
		idx  map[string]*protocol.CardView
		seen map[cardKey]bool
	)
	seat := ev.Seat.String()
	keyOf := func(m legal.Move) (cardKey, *protocol.CardView, bool) {
		action := cardAction(m.Kind)
		if action == "" || m.Source == uuid.Nil {
			return cardKey{}, nil, false
		}
		if idx == nil {
			idx = indexView(ev.Input.View)
		}
		c := idx[m.Source.String()]
		if c == nil || c.IsToken || c.Owner != seat || isLand(c.TypeLine) || c.Name == "" {
			return cardKey{}, nil, false
		}
		return cardKey{c.Name, action}, c, true
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, m := range ev.Input.Moves {
		k, c, ok := keyOf(m)
		if !ok || seen[k] {
			continue
		}
		if seen == nil {
			seen = map[cardKey]bool{}
		}
		seen[k] = true
		u := t.uses[k]
		if u == nil {
			u = &CardUse{Name: k.name, Action: k.action}
			t.uses[k] = u
		}
		u.ManaSource = u.ManaSource || repeatableManaSource(c)
		u.Offered++
	}
	if !ev.Applied || ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	if k, _, ok := keyOf(ev.Input.Moves[ev.Index]); ok {
		if u := t.uses[k]; u != nil {
			u.Taken++
		}
	}
}

// list is the tally in a stable order.
func (t *cardTally) list() []CardUse {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]CardUse, 0, len(t.uses))
	for _, u := range t.uses {
		out = append(out, *u)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Action < out[j].Action
	})
	return out
}

// repeatableManaSource is ADR 0126 §2's "repeatable" mana source: a
// tap mana ability with no sacrifice or exile in its cost.
func repeatableManaSource(c *protocol.CardView) bool {
	for _, m := range c.ManaAbilities {
		if m.TapCost && !m.SacrificeCost && !m.ExileSelf && m.SacrificeLabel == "" && m.ExilePermanentLabel == "" {
			return true
		}
	}
	return false
}

func cardAction(k legal.Kind) string {
	switch k {
	case legal.KindCast:
		return ActionCast
	case legal.KindActivate:
		return ActionActivate
	}
	return ""
}

// isLand reads the front face of a type line: a modal double-faced
// card whose front is a spell is a spell.
func isLand(typeLine string) bool {
	front, _, _ := strings.Cut(typeLine, "//")
	return strings.Contains(front, "Land")
}

// indexView maps instance ID to card over every zone a seat can cast
// or activate from.
func indexView(v protocol.GameView) map[string]*protocol.CardView {
	out := map[string]*protocol.CardView{}
	add := func(z *protocol.ZoneView) {
		for i := range z.Cards {
			if id := z.Cards[i].InstanceID; id != "" {
				out[id] = &z.Cards[i]
			}
		}
	}
	add(&v.Battlefield)
	add(&v.Exile)
	for i := range v.Seats {
		s := &v.Seats[i]
		add(&s.Hand)
		add(&s.Command)
		add(&s.Graveyard)
		add(&s.Library)
	}
	return out
}

// CardTotals is one card across a contestant's whole run.
type CardTotals struct {
	Name       string `json:"name"`
	Action     string `json:"action"`
	ManaSource bool   `json:"mana_source,omitempty"`
	// Windows and Taken are summed over every window of every game.
	Windows int `json:"windows"`
	Taken   int `json:"taken"`
	// GamesOffered and GamesUsed are SEAT-games: the seat-games in
	// which the card was offered at least once, and those in which it
	// was used at least once. With one seat per contestant per game,
	// as in a four-deck rotation, they are games.
	GamesOffered int `json:"games_offered"`
	GamesUsed    int `json:"games_used"`
	// Never is Windows >= NeverWindows and Taken == 0.
	Never bool `json:"never,omitempty"`
}

// UseRate is GamesUsed / GamesOffered, 0 when never offered.
func (c CardTotals) UseRate() float64 {
	if c.GamesOffered == 0 {
		return 0
	}
	return float64(c.GamesUsed) / float64(c.GamesOffered)
}

// ContestantCards is the Cards section for one contestant.
type ContestantCards struct {
	Policy string `json:"policy"`
	Deck   string `json:"deck,omitempty"`
	// SeatGames is how many seat-games the contestant played.
	SeatGames int          `json:"seat_games"`
	Cards     []CardTotals `json:"cards"`
	// Never is how many of Cards are flagged never.
	Never int `json:"never"`
}

// Contestant names it the way the report does.
func (c ContestantCards) Contestant() string { return contestantName(c.Policy, c.Deck) }

func contestantName(policy, deck string) string {
	if deck == "" {
		return policy
	}
	return policy + " · " + deck
}

// cardsAcc folds per-seat tallies into ContestantCards.
type cardsAcc struct {
	per map[contestantKey]*cardsOf
}

type contestantKey struct{ policy, deck string }

type cardsOf struct {
	seatGames int
	cards     map[cardKey]*CardTotals
}

func newCardsAcc() *cardsAcc { return &cardsAcc{per: map[contestantKey]*cardsOf{}} }

func (a *cardsAcc) add(r GameResult) {
	for _, s := range r.Seats {
		k := contestantKey{s.Spec.Label(), s.Spec.Deck}
		c := a.per[k]
		if c == nil {
			c = &cardsOf{cards: map[cardKey]*CardTotals{}}
			a.per[k] = c
		}
		c.seatGames++
		for _, u := range s.Cards {
			ck := cardKey{u.Name, u.Action}
			t := c.cards[ck]
			if t == nil {
				t = &CardTotals{Name: u.Name, Action: u.Action}
				c.cards[ck] = t
			}
			t.ManaSource = t.ManaSource || u.ManaSource
			t.Windows += u.Offered
			t.Taken += u.Taken
			if u.Offered > 0 {
				t.GamesOffered++
			}
			if u.Taken > 0 {
				t.GamesUsed++
			}
		}
	}
}

func (a *cardsAcc) totals() []ContestantCards {
	out := make([]ContestantCards, 0, len(a.per))
	for k, c := range a.per {
		cc := ContestantCards{Policy: k.policy, Deck: k.deck, SeatGames: c.seatGames, Cards: make([]CardTotals, 0, len(c.cards))}
		for _, t := range c.cards {
			v := *t
			v.Never = v.Windows >= NeverWindows && v.Taken == 0
			if v.Never {
				cc.Never++
			}
			cc.Cards = append(cc.Cards, v)
		}
		sort.Slice(cc.Cards, func(i, j int) bool {
			a, b := cc.Cards[i], cc.Cards[j]
			if a.Name != b.Name {
				return a.Name < b.Name
			}
			return a.Action < b.Action
		})
		out = append(out, cc)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Policy != out[j].Policy {
			return out[i].Policy < out[j].Policy
		}
		return out[i].Deck < out[j].Deck
	})
	return out
}

// Canary is a card ADR 0126 names in its acceptance bar: one of the
// six A3 canaries, each standing for a class of card the heuristic
// never played.
type Canary struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	Class  string `json:"class"`
}

// Canaries are ADR 0126's A3 cards.
var Canaries = []Canary{
	{Name: "Sol Ring", Action: ActionCast, Class: "mana rock"},
	{Name: "Rhystic Study", Action: ActionCast, Class: "enchantment engine"},
	{Name: "Mary Read and Anne Bonny", Action: ActionActivate, Class: "loot"},
	{Name: "Entomb", Action: ActionCast, Class: "tutor"},
	{Name: "Harrow", Action: ActionCast, Class: "ramp spell"},
	{Name: "Viscera Seer", Action: ActionCast, Class: "sacrifice outlet"},
}

// The bars ADR 0126 sets on the share of games a card was used in,
// out of the games it was offered in.
const (
	// ManaSourceBar is A2's: every mana rock and dork.
	ManaSourceBar = 0.80
	// CanaryBar is A3's: each of the six canaries.
	CanaryBar = 0.50
)

// CanaryResult is one acceptance-bar card for one contestant.
type CanaryResult struct {
	// Bar is "A2" (a mana rock or dork) or "A3" (a named canary).
	Bar        string  `json:"bar"`
	Contestant string  `json:"contestant"`
	Name       string  `json:"name"`
	Action     string  `json:"action"`
	Class      string  `json:"class"`
	Offered    int     `json:"games_offered"`
	Used       int     `json:"games_used"`
	Rate       float64 `json:"rate"`
	Want       float64 `json:"want"`
	Meets      bool    `json:"meets"`
}

// canaries reads A2 and A3 off the Cards section. An A3 canary that no
// contestant was offered gets one row with no contestant, so the
// report says it was not measured rather than leaving it out.
func canaries(cards []ContestantCards) []CanaryResult {
	var a2, a3 []CanaryResult
	seen := map[cardKey]bool{}
	row := func(bar string, cc ContestantCards, t CardTotals, class string, want float64) CanaryResult {
		return CanaryResult{
			Bar: bar, Contestant: cc.Contestant(), Name: t.Name, Action: t.Action, Class: class,
			Offered: t.GamesOffered, Used: t.GamesUsed, Rate: t.UseRate(), Want: want,
			Meets: t.GamesOffered > 0 && t.UseRate() >= want,
		}
	}
	for _, cc := range cards {
		for _, t := range cc.Cards {
			if t.ManaSource && t.Action == ActionCast {
				a2 = append(a2, row("A2", cc, t, "mana rock or dork", ManaSourceBar))
			}
		}
	}
	for _, cn := range Canaries {
		for _, cc := range cards {
			for _, t := range cc.Cards {
				if t.Name == cn.Name && t.Action == cn.Action {
					a3 = append(a3, row("A3", cc, t, cn.Class, CanaryBar))
					seen[cardKey{cn.Name, cn.Action}] = true
				}
			}
		}
		if !seen[cardKey{cn.Name, cn.Action}] {
			a3 = append(a3, CanaryResult{Bar: "A3", Name: cn.Name, Action: cn.Action, Class: cn.Class, Want: CanaryBar})
		}
	}
	return append(a2, a3...)
}

// writeCards renders the Cards section.
func writeCards(b *strings.Builder, s Summary) {
	if len(s.Cards) == 0 {
		return
	}
	b.WriteString("\n### Cards\n\n")
	fmt.Fprintf(b, "Every non-land card each contestant was offered: a `cast` or `activate` move naming it. `never` is offered in %d or more windows and taken in none. Games are seat-games.\n\n", NeverWindows)
	b.WriteString("| contestant | seat-games | cards offered | never | never cast or used |\n")
	b.WriteString("|---|---:|---:|---:|---|\n")
	for _, cc := range s.Cards {
		var never []string
		for _, t := range cc.Cards {
			if t.Never {
				never = append(never, cardName(t))
			}
		}
		fmt.Fprintf(b, "| %s | %d | %d | %d | %s |\n", cc.Contestant(), cc.SeatGames, len(cc.Cards), cc.Never, orDash(strings.Join(never, ", ")))
	}

	if len(s.Canaries) > 0 {
		b.WriteString("\n#### Acceptance-bar cards (ADR 0126 A2, A3)\n\n")
		b.WriteString("| bar | card | class | contestant | games offered | games used | used | want | meets |\n")
		b.WriteString("|---|---|---|---|---:|---:|---:|---:|:--:|\n")
		for _, c := range s.Canaries {
			if c.Contestant == "" {
				fmt.Fprintf(b, "| %s | %s | %s | — | 0 | 0 | — | %s | not offered |\n", c.Bar, cardName(CardTotals{Name: c.Name, Action: c.Action}), c.Class, pct0(c.Want))
				continue
			}
			fmt.Fprintf(b, "| %s | %s | %s | %s | %d | %d | %s | %s | %s |\n",
				c.Bar, cardName(CardTotals{Name: c.Name, Action: c.Action}), c.Class, c.Contestant,
				c.Offered, c.Used, pct0(c.Rate), pct0(c.Want), yesNo(c.Meets))
		}
	}

	for _, cc := range s.Cards {
		fmt.Fprintf(b, "\n<details><summary>%s: %d cards offered, %d never</summary>\n\n", cc.Contestant(), len(cc.Cards), cc.Never)
		b.WriteString("| card | action | mana | windows offered | taken | games offered | games used | never |\n")
		b.WriteString("|---|---|:--:|---:|---:|---:|---:|:--:|\n")
		for _, t := range cc.Cards {
			mana := ""
			if t.ManaSource {
				mana = "yes"
			}
			never := ""
			if t.Never {
				never = "**never**"
			}
			fmt.Fprintf(b, "| %s | %s | %s | %d | %d | %d | %d | %s |\n",
				t.Name, t.Action, mana, t.Windows, t.Taken, t.GamesOffered, t.GamesUsed, never)
		}
		b.WriteString("\n</details>\n")
	}
}

// cardName is a card as the report names it: an activation says so.
func cardName(t CardTotals) string {
	if t.Action == ActionActivate {
		return t.Name + " (activate)"
	}
	return t.Name
}

func pct0(v float64) string { return fmt.Sprintf("%.0f%%", v*100) }
