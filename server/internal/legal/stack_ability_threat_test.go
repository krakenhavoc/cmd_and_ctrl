package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// stack_ability_threat_test.go — #4779. The engine keeps an activated or
// triggered ability on the stack as metadata alone (g.StackMeta), with no
// card on the stack zone. stackThreatens and selfSacrificeThreatened read
// only the spell cards in g.Stack, so an opponent's ability aimed at the
// seat's creature set neither a threat-gated row's interacts nor a
// self-sacrifice row's, while answers_now (#4734), which reads StackMeta,
// said yes. Royal Assassin's activation and Ravenous Chupacabra's enters
// trigger are the opponent's items; Undercity Troll (survive, gated) and
// Sakura-Tribe Elder (sacrifice itself) are the seat's answers.

const (
	oracleRoyalAssassinSA      = "9ed6f28f-a3db-48c5-9ab0-b90a7fba5f57"
	oracleRavenousChupacabraSA = "7b459306-149b-4f43-abc1-2dd70c748c0e"
)

var (
	royalAssassinSA = game.Card{
		Name: "Royal Assassin", TypeLine: "Creature — Human Assassin", OracleID: oracleRoyalAssassinSA,
		ManaCost: "{1}{B}{B}", Power: 1, Toughness: 1,
	}
	ravenousChupacabraSA = game.Card{
		Name: "Ravenous Chupacabra", TypeLine: "Creature — Beast Horror", OracleID: oracleRavenousChupacabraSA,
		ManaCost: "{2}{B}{B}", Power: 2, Toughness: 2,
	}
)

// pushAimed puts `by`'s item from `source` on the stack aimed at
// `victim`: Royal Assassin's activation or Ravenous Chupacabra's trigger.
type pushAimed func(t *testing.T, b *threatBoard, by *game.Player, victim uuid.UUID)

var stackAbilitiesSA = []struct {
	name string
	push pushAimed
}{
	{"an activated ability", func(t *testing.T, b *threatBoard, by *game.Player, victim uuid.UUID) {
		source := battlefieldCard(b.g, by, royalAssassinSA)
		b.pushActivated(t, by, source, "Destroy target tapped creature", game.TargetRef{Kind: game.TargetCard, ID: victim})
	}},
	{"a triggered ability", func(t *testing.T, b *threatBoard, by *game.Player, victim uuid.UUID) {
		source := battlefieldCard(b.g, by, ravenousChupacabraSA)
		b.pushTriggered(t, by, source, "destroy target creature", game.TargetRef{Kind: game.TargetCard, ID: victim})
	}},
}

// assertNoStackCards fails when an ability left a card on the stack zone:
// the probe is only meaningful when the item is metadata alone.
func assertNoStackCards(t *testing.T, b *threatBoard) {
	t.Helper()
	b.g.WithWriteLock(func() {
		if n := len(b.g.Stack.Cards); n != 0 {
			t.Fatalf("the stack zone holds %d cards; an ability should be metadata alone", n)
		}
	})
}

// answersNow is the move's answers_now verdict, false when it has none.
func answersNow(m legal.Move) bool { return m.AnswersNow != nil && *m.AnswersNow }

func TestAnOpponentsAbilityAimedAtYourCreatureThreatensYou(t *testing.T) {
	for _, tc := range stackAbilitiesSA {
		t.Run(tc.name+" aimed at Undercity Troll", func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			tc.push(t, b, b.opp, b.troll)
			assertNoStackCards(t, b)
			// Every gated answer of the seat's stops, the Troll's included,
			// and agrees with answers_now.
			b.checkGated(t, true)
			for _, m := range movesFrom(legal.EnumerateFor(b.g, b.seat.ID), b.troll, legal.KindActivate) {
				if !answersNow(m) {
					t.Errorf("Undercity Troll: answers_now=%v, want true (%s)", m.AnswersNow, m.Label)
				}
			}
		})
		t.Run(tc.name+" aimed at the opponent's own creature", func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			tc.push(t, b, b.opp, b.oppBears)
			assertNoStackCards(t, b)
			b.checkGated(t, false)
		})
		t.Run("the seat's own "+tc.name+" aimed at its own creature", func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			tc.push(t, b, b.seat, b.bears)
			assertNoStackCards(t, b)
			b.checkGated(t, false)
		})
	}
}

func TestAnOpponentsAbilityAimedAtASelfSacrificingCreatureMakesItAnAnswer(t *testing.T) {
	for _, tc := range stackAbilitiesSA {
		t.Run(tc.name, func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			elder := battlefieldCard(b.g, b.seat, game.Card{
				Name: "Sakura-Tribe Elder", TypeLine: "Creature — Snake Shaman", ManaCost: "{1}{G}",
				OracleID: oracleSakuraTribeElderAS, Power: 1, Toughness: 1,
			})
			if i, c := elderFlags(t, b.g, b.seat, elder); i || c {
				t.Fatalf("an unthreatened Elder: interacts=%v combat_interacts=%v, want neither", i, c)
			}
			tc.push(t, b, b.opp, elder)
			assertNoStackCards(t, b)
			if i, c := elderFlags(t, b.g, b.seat, elder); !i || c {
				t.Errorf("an Elder an opponent's ability targets: interacts=%v combat_interacts=%v, want interacts only", i, c)
			}
		})
		t.Run("the seat's own "+tc.name, func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			elder := battlefieldCard(b.g, b.seat, game.Card{
				Name: "Sakura-Tribe Elder", TypeLine: "Creature — Snake Shaman", ManaCost: "{1}{G}",
				OracleID: oracleSakuraTribeElderAS, Power: 1, Toughness: 1,
			})
			tc.push(t, b, b.seat, elder)
			if i, _ := elderFlags(t, b.g, b.seat, elder); i {
				t.Error("an Elder its own controller's ability targets interacts")
			}
		})
	}
}
