package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

// opening_roll_kat_test.go pins what StartWithFirstPlayerRoll produced
// BEFORE ADR 0121 moved the shuffle and the deal after the choice and
// rebuilt the automatic path on the opening-roll window (§1, "the same
// winner either way"). The expected values below were captured from the
// pre-0121 engine (origin/develop at dc4f1844) and must never be
// regenerated: a seeded arena game, the probe and the demo seed are
// promised the same rolls, the same starting seat and the same opening
// hands and libraries as before the change.
//
// What is pinned per table: the starting seat, every opening die in the
// order it was rolled (seat:result), and a digest of every seat's hand
// and library, in order, by card name. The order of those dice relative
// to anything else in the event log is deliberately NOT pinned: the
// dice now come before the deal, where CR 103.3 and 903.7 put them.

type openingRollKAT struct {
	seats  int
	seed1  uint64
	seed2  uint64
	winner int
	rolls  string
	zones  string // hex sha256 of openingZonesDigest
}

var openingRollKnownAnswers = []openingRollKAT{
	{seats: 2, seed1: 1, seed2: 2027, winner: 0, rolls: "0:20,1:17", zones: "559b290ce3535e4e4948fa3ea01e469d754744d09c00a3b307df57b5464caa6a"},
	{seats: 2, seed1: 2, seed2: 2028, winner: 1, rolls: "0:1,1:6", zones: "1c192cc59d5afe532912042bdb859e6c221c48a77f29203be8e1353f4e8cf193"},
	{seats: 2, seed1: 3, seed2: 2029, winner: 1, rolls: "0:10,1:20", zones: "c09d18545c97eecb98d4ea0e0c17407faa71b168c7e07076eae5389e5bda9b91"},
	{seats: 2, seed1: 4, seed2: 2030, winner: 0, rolls: "0:20,1:15", zones: "9f6de245cfc8af21391f3e453dbca37da10be3f246be6adbeacccb4ce9a6b7ff"},
	{seats: 2, seed1: 5, seed2: 2031, winner: 0, rolls: "0:15,1:12", zones: "7d8d6f6c79a29d830f4043cabd1dceb5910d71d93bdfaac2192416fdc8d12163"},
	{seats: 2, seed1: 6, seed2: 2032, winner: 0, rolls: "0:19,1:6", zones: "2ab821a9b5f461c776c078594c30d92e6a1751622e4c5666eb960e74c2be78fc"},
	{seats: 2, seed1: 7, seed2: 2033, winner: 1, rolls: "0:15,1:16", zones: "d3a1df5bad6539fd66277ee459c7a8729d8676d9bea05671dbc241ec005965c4"},
	{seats: 2, seed1: 8, seed2: 2034, winner: 0, rolls: "0:18,1:13", zones: "40d81efb6442d994773559e70999d4daf538b1783f834540dc5247609373d082"},
	{seats: 3, seed1: 1, seed2: 2027, winner: 0, rolls: "0:20,1:17,2:1", zones: "29a1c6ef9274901c3d0e24af4bfa7b936e6031c078ead55471dae8eba79caa8d"},
	{seats: 3, seed1: 2, seed2: 2028, winner: 1, rolls: "0:1,1:6,2:5", zones: "d4c093a96360ef63fb71e9a552d081929ce631e06b79205b1c3ecdcd97c1c7db"},
	{seats: 3, seed1: 3, seed2: 2029, winner: 1, rolls: "0:10,1:20,2:7", zones: "0022cca61ebb3a5c9ee17f11776f31f2f961eed7d26d50646868600048ff0111"},
	{seats: 3, seed1: 4, seed2: 2030, winner: 0, rolls: "0:20,1:15,2:12", zones: "212cdd863c5c22b9098eb72b8bbe3afd1a6bde9996a522d773d1f8316787253c"},
	{seats: 3, seed1: 5, seed2: 2031, winner: 0, rolls: "0:15,1:12,2:14", zones: "46a656fae2e74bb453c21907148f02e4aee47a6aed471147809cb1d7f28b1513"},
	{seats: 3, seed1: 6, seed2: 2032, winner: 0, rolls: "0:19,1:6,2:9", zones: "09ff54e05b232780e65e59e65b2c3ac32603e42ff4191631f2993e45ae533a46"},
	{seats: 3, seed1: 7, seed2: 2033, winner: 1, rolls: "0:15,1:16,2:2", zones: "911f14841216b1addae7fd3cb8a61b1a586ac08829fb44761c9cb0e7931c9c2d"},
	{seats: 3, seed1: 8, seed2: 2034, winner: 0, rolls: "0:18,1:13,2:5", zones: "bd89bcd1e89b2c40380ca5e83eb65581c277003a5f3aca53609113c458b86ddd"},
	{seats: 4, seed1: 1, seed2: 2027, winner: 0, rolls: "0:20,1:17,2:1,3:10", zones: "75196cb85bcc1d9b857186f47bcbeed9c897bd41ae65d7f0ab0a53db3f9d07a1"},
	{seats: 4, seed1: 2, seed2: 2028, winner: 3, rolls: "0:1,1:6,2:5,3:15", zones: "7d13694b45bdae3b5d1a0171f0dbb09196cf2725b744b25c63e9ab4caf5c7a3b"},
	{seats: 4, seed1: 3, seed2: 2029, winner: 1, rolls: "0:10,1:20,2:7,3:19", zones: "80ccf921d88ae2a88c2be7a0d312c190b3549f959b879b4a6f28bd9771001899"},
	{seats: 4, seed1: 4, seed2: 2030, winner: 0, rolls: "0:20,1:15,2:12,3:3", zones: "f18388cb5ecdcc859b9cac73bb31a7d2489afe67857e49b2724e6da5c13c6316"},
	{seats: 4, seed1: 5, seed2: 2031, winner: 3, rolls: "0:15,1:12,2:14,3:17", zones: "4ce8bb1d7ee6fc6583ff62ad15b8fe045999e031dde0deee18b25c505a6619a5"},
	{seats: 4, seed1: 6, seed2: 2032, winner: 0, rolls: "0:19,1:6,2:9,3:4", zones: "6ea0778e9cc38eee424e1b86dd45886d65b16dcc4f51d77ce976c3107e36cec7"},
	{seats: 4, seed1: 7, seed2: 2033, winner: 1, rolls: "0:15,1:16,2:2,3:3", zones: "4ed2a4bb3b1a044ead9bd373d2c783671bccb2f558a698a0af82d459054090a7"},
	{seats: 4, seed1: 8, seed2: 2034, winner: 0, rolls: "0:18,1:13,2:5,3:2", zones: "c8035b665831d9bb06a64b3253c3270f2dc82a4690499da7b009d79d7878bf82"},
	// Tables whose opening roll tied, captured the same way.
	{seats: 2, seed1: 10, seed2: 2036, winner: 0, rolls: "0:10,1:10,0:20,1:10", zones: "97002ae4fa96186aed01612efa9cbd4e0c58e4c81a8a42132bf409607b646cfa"},
	{seats: 2, seed1: 15, seed2: 2041, winner: 0, rolls: "0:7,1:7,0:18,1:18,0:1,1:1,0:10,1:1", zones: "96c51aa90a178c39c5c9514e882688a13625d11b120d2a8b72f5dd60d032fce6"},
	{seats: 3, seed1: 15, seed2: 2041, winner: 0, rolls: "0:7,1:7,2:7,0:18,1:18,2:14,0:1,1:1,0:10,1:1", zones: "d965fd2829967bfbf117bbb19ef43a2eca0fc1455695d2f940144b847a9de946"},
	{seats: 3, seed1: 17, seed2: 2043, winner: 0, rolls: "0:6,1:3,2:6,0:7,2:2", zones: "603e317a4235ca216081a421398c02d7a371223ab5381a583eb712829447c72a"},
	{seats: 3, seed1: 184, seed2: 2210, winner: 2, rolls: "0:6,1:12,2:12,1:2,2:2,1:7,2:11", zones: "a166dc9f5fa6e418aaf2e9f84b6638bdbfb91e04c475cee56c88629f7eb50695"},
	{seats: 4, seed1: 10, seed2: 2036, winner: 0, rolls: "0:10,1:10,2:9,3:5,0:20,1:10", zones: "77ff2fd4ad30eacd3ac8b1911854b4b8d1848c07cf538162d46aebbfbf438ef0"},
	{seats: 4, seed1: 22, seed2: 2048, winner: 3, rolls: "0:7,1:4,2:8,3:8,2:9,3:18", zones: "fca040bc797dd129f499da9e1efa9b9e75a6731a32f4463f31c866207d440b44"},
	{seats: 4, seed1: 23, seed2: 2049, winner: 2, rolls: "0:19,1:3,2:20,3:20,2:19,3:14", zones: "def3837f3262ec137a035df1dfa93dde9235cb000a3452b52f75e5aff3114ee3"},
	{seats: 4, seed1: 50, seed2: 2076, winner: 0, rolls: "0:20,1:1,2:2,3:20,0:11,3:3", zones: "673b6780e3391ece744d4e2e5344749f13f7d1b0c195a89faca0db52e7d47840"},
	{seats: 4, seed1: 113, seed2: 2139, winner: 2, rolls: "0:16,1:18,2:18,3:5,1:7,2:20", zones: "2c9a26f94de06ddbeb901aa86fc3e2843c1fb1c98b880bd5846f7a8b0fcb63db"},
}

// startAutomaticForKAT seats n players with distinct decks and starts
// them through the automatic opening roll on the seeded source.
func startAutomaticForKAT(t *testing.T, n int, seed1, seed2 uint64) *Game {
	t.Helper()
	g := NewGame()
	for seat := 0; seat < n; seat++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", seat+1), buildTestDeck(fmt.Sprintf("Commander %d", seat+1))); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithFirstPlayerRoll(rand.New(rand.NewPCG(seed1, seed2))); err != nil {
		t.Fatalf("StartWithFirstPlayerRoll: %v", err)
	}
	return g
}

// openingRollsOf lists every opening die in the event log as
// "seat:result", in emit order.
func openingRollsOf(g *Game) string {
	var parts []string
	for _, ev := range g.Events {
		if ev.Kind != EventRollDie {
			continue
		}
		seat := -1
		for i, p := range g.Seats {
			if p.ID == ev.Actor {
				seat = i
			}
		}
		parts = append(parts, fmt.Sprintf("%d:%d", seat, ev.Amount))
	}
	return strings.Join(parts, ",")
}

// openingZonesDigest hashes every seat's hand and library, in order,
// by card name.
func openingZonesDigest(g *Game) string {
	var b strings.Builder
	for i, p := range g.Seats {
		fmt.Fprintf(&b, "seat %d hand:", i)
		for _, c := range p.Hand.Cards {
			b.WriteString(c.Name + "|")
		}
		b.WriteString(" library:")
		for _, c := range p.Library.Cards {
			b.WriteString(c.Name + "|")
		}
		b.WriteString("\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func TestStartWithFirstPlayerRollKnownAnswers(t *testing.T) {
	if len(openingRollKnownAnswers) == 0 {
		// Capture mode: print the table for the file above.
		for _, n := range []int{2, 3, 4} {
			for s := uint64(1); s <= 8; s++ {
				g := startAutomaticForKAT(t, n, s, 2026+s)
				t.Logf("{seats: %d, seed1: %d, seed2: %d, winner: %d, rolls: %q, zones: %q},",
					n, s, 2026+s, g.StartingSeat, openingRollsOf(g), openingZonesDigest(g))
			}
		}
		t.Fatal("no known answers recorded")
	}
	for _, kat := range openingRollKnownAnswers {
		t.Run(fmt.Sprintf("%dseats/%d", kat.seats, kat.seed1), func(t *testing.T) {
			g := startAutomaticForKAT(t, kat.seats, kat.seed1, kat.seed2)
			if g.StartingSeat != kat.winner {
				t.Errorf("starting seat = %d, want %d", g.StartingSeat, kat.winner)
			}
			if g.Turn.ActiveSeat != kat.winner || g.Turn.Seq != 1 {
				t.Errorf("turn = %+v, want seat %d at Seq 1", g.Turn, kat.winner)
			}
			if got := openingRollsOf(g); got != kat.rolls {
				t.Errorf("opening dice = %s, want %s", got, kat.rolls)
			}
			if got := openingZonesDigest(g); got != kat.zones {
				t.Errorf("hands and libraries digest = %s, want %s", got, kat.zones)
			}
			for i, p := range g.Seats {
				if len(p.Hand.Cards) != OpeningHandSize {
					t.Errorf("seat %d hand = %d cards, want %d", i, len(p.Hand.Cards), OpeningHandSize)
				}
			}
		})
	}
}
