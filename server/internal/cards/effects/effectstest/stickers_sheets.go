package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stkScene is a table with one sheet, a creature of the seat's and one
// ability sticker on it.
type StkScene struct {
	t    *testing.T
	G    *game.Game
	Me   *game.Player
	Opp  *game.Player
	Host uuid.UUID
}

// newStkScene puts the ability sticker whose label contains `sticker` from
// `sheet` on a 2/2 Bear the seat owns.
func NewStkScene(t *testing.T, sheet, sticker string) *StkScene {
	t.Helper()
	return NewStkSceneOn(t, sheet, sticker, "Creature — Bear", 2, 2)
}

// newStkSceneOn is newStkScene with the host's type line and power and
// toughness chosen.
func NewStkSceneOn(t *testing.T, sheet, sticker, typeLine string, power, toughness int) *StkScene {
	t.Helper()
	g, me := StkTableWith(t, 12, sheet)
	s := &StkScene{t: t, G: g, Me: me, Opp: g.Seats[1]}
	s.Host = P2Push(g, me.ID, "Host", typeLine, "", power, toughness)
	StkGive(t, g, me, s.Host, sticker)
	return s
}

func (s *StkScene) Settle() { s.t.Helper(); PassPriorityAroundTable(s.t, s.G) }

// attack declares the host as an attacker at the opponent and settles the
// attack triggers (stopping at a prompt).
func (s *StkScene) Attack() {
	s.t.Helper()
	DeclareAttack(s.t, s.G, s.Opp.ID, s.Host)
	s.Settle()
}

// activate activates the host's ability `index` with the mana named.
func (s *StkScene) Activate(index int, mana string, targets ...game.TargetRef) {
	s.t.Helper()
	if err := P2Activate(s.t, s.G, s.Me, s.Host, index, mana, targets...); err != nil {
		s.t.Fatalf("activate %d: %v", index, err)
	}
	s.Settle()
}

// destroyHost destroys the host and settles what that triggers.
func (s *StkScene) DestroyHost() {
	s.t.Helper()
	B25Destroy(s.G, s.Host)
	s.Settle()
}

// pickTargets answers the open target question.
func (s *StkScene) PickTargets(targets ...game.TargetRef) {
	s.t.Helper()
	p := LatestPickTarget(s.G, s.Me.ID)
	if p == nil {
		s.t.Fatal("no target question is open")
	}
	if err := s.G.ResolvePickTargets(p.ID, s.Me.ID, targets); err != nil {
		s.t.Fatalf("pick targets: %v", err)
	}
	s.Settle()
}

func (s *StkScene) Card() game.Card { s.t.Helper(); return StkCard(s.t, s.G, s.Host) }

func (s *StkScene) Has(kw string) bool { return P2Keyword(s.G, s.Host, kw) }
