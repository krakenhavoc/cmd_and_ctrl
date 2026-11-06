package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// team_pump_test.go — #4320: an untargeted pump of every creature the bot
// controls until end of turn is worth the attack it adds, not the flat
// ActivateBase. Purphoros, God of the Forge pumped a team that never
// attacked, before combat and after it.

// purphorosGod is a seat-0 Purphoros below its devotion, so not a creature,
// with its one activated row as the catalog declares it.
func purphorosGod(id string) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id, Name: "Purphoros, God of the Forge", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Legendary Enchantment — God", ManaCost: "{3}{R}", KnownByYou: true,
	}
	answers := []string{"pump"}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{2}{R}: Creatures you control get +1/+0 until end of turn.", ManaCost: "{2}{R}",
		Purpose: &protocol.PurposeView{Answers: &answers, TeamPump: &protocol.PumpView{Power: 1}},
	}}
	return c
}

func purphorosPump(t *testing.T, src string) legal.Move {
	return activateMove(t, 0, src, "Purphoros: creatures you control get +1/+0", &legal.MoveCost{Mana: "{2}{R}"})
}

// teamPumpBoard is six untapped Mountains and Purphoros, plus extra.
func teamPumpBoard(src string, extra ...protocol.CardView) []protocol.CardView {
	return append(append(manaLands(6, 0, 90), purphorosGod(src)), extra...)
}

func noTeamPumps() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceTeamPumps = false
	return heuristic.NewWithConfig(cfg)
}

// TestTeamPumpWaitsForAnAttack: the bot's first main phase, six mana and
// nothing to cast. With no creature that can attack this turn (none, or
// only a summoning-sick one) the pump adds nothing and the bot passes;
// the back-out keeps the flat price and pumps, as the issue recorded.
func TestTeamPumpWaitsForAnAttack(t *testing.T) {
	src := cardID(1)
	boards := map[string][]protocol.CardView{
		"no creature":             teamPumpBoard(src),
		"a summoning-sick 2/2":    teamPumpBoard(src, creature(cardID(3), 0, "Fresh Bear", 2, 2, sick())),
		"a tapped 3/3":            teamPumpBoard(src, creature(cardID(3), 0, "Tired Ogre", 3, 3, tapped())),
		"an opponent's 5/5 alone": teamPumpBoard(src, creature(cardID(4), 1, "Their Wurm", 5, 5)),
	}
	for name, board := range boards {
		t.Run(name, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, "precombat_main"))
			moves := []legal.Move{passMove(0), purphorosPump(t, src)}
			if i, why := decideOn(t, heuristic.New(), v, moves...); i != 0 {
				t.Errorf("the bot pumped a team that cannot attack: %s", why)
			}
			if got := rankValue(t, heuristic.New(), input(0, v, moves...), moves[1].Label); got > 0 {
				t.Errorf("the pump is priced %g, want nothing", got)
			}
			if i, _ := decideOn(t, noTeamPumps(), v, moves...); i != 1 {
				t.Errorf("the back-out passed; it is meant to keep the flat ActivateBase")
			}
		})
	}
}

// TestTeamPumpIsTakenForItsAttackers: the bot's declare attackers step,
// two 2/2s attacking an opponent with no blockers, six mana and nothing
// in hand. Each +1/+0 is a point of damage more, and the mana is spare,
// so the bot pumps; the reason names the spare-mana bar.
func TestTeamPumpIsTakenForItsAttackers(t *testing.T) {
	src := cardID(1)
	board := teamPumpBoard(src,
		creature(cardID(3), 0, "Attacker A", 2, 2, attacking(1), tapped()),
		creature(cardID(4), 0, "Attacker B", 2, 2, attacking(1), tapped()))
	for _, step := range []string{"declare_attackers", "declare_blockers"} {
		t.Run(step, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, step))
			i, why := decideOn(t, heuristic.New(), v, passMove(0), purphorosPump(t, src))
			if i != 1 {
				t.Fatalf("picked %d (%s), want the pump for two attackers", i, why)
			}
			if !strings.Contains(why, "spare mana in combat") {
				t.Errorf("reason %q, want the spare-mana bar", why)
			}
		})
	}
}

// TestTeamPumpKeepsTheMainPhaseMana: the same attack with a {2}{R}
// creature in hand and three Mountains. The pump would spend the mana the
// second main phase casts with, so it does not clear the response bar.
func TestTeamPumpKeepsTheMainPhaseMana(t *testing.T) {
	src := cardID(1)
	bear := creature(cardID(5), 0, "Hand Bear", 3, 3)
	bear.ManaCost = "{2}{R}"
	board := append(append(manaLands(3, 0, 90), purphorosGod(src)),
		creature(cardID(3), 0, "Attacker A", 2, 2, attacking(1), tapped()),
		creature(cardID(4), 0, "Attacker B", 2, 2, attacking(1), tapped()))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(bear)), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, "declare_attackers"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), purphorosPump(t, src)); i != 0 {
		t.Errorf("the bot spent its main-phase mana on a pump: %s", why)
	}
}

// TestTeamPumpAfterDamageIsWorthNothing: after combat damage the pump has
// nothing left to change, in the end of combat step with the attackers
// still marked and in the second main phase.
func TestTeamPumpAfterDamageIsWorthNothing(t *testing.T) {
	src := cardID(1)
	for _, step := range []string{"combat_damage", "end_combat", "postcombat_main"} {
		t.Run(step, func(t *testing.T) {
			board := teamPumpBoard(src,
				creature(cardID(3), 0, "Attacker A", 2, 2, attacking(1), tapped()),
				creature(cardID(4), 0, "Attacker B", 2, 2, attacking(1), tapped()),
				creature(cardID(6), 0, "Home Guard", 2, 2))
			if step == "postcombat_main" {
				for i := range board {
					board[i].AttackingTarget = ""
				}
			}
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, step))
			moves := []legal.Move{passMove(0), purphorosPump(t, src)}
			if i, why := decideOn(t, heuristic.New(), v, moves...); i != 0 {
				t.Errorf("the bot pumped after combat damage: %s", why)
			}
			if got := rankValue(t, heuristic.New(), input(0, v, moves...), moves[1].Label); got > 0 {
				t.Errorf("the pump is priced %g after damage, want nothing", got)
			}
		})
	}
}

// TestTeamHasteLetsFreshCreaturesAttack: Goro-Goro's "{R}: Creatures you
// control gain haste until end of turn." before combat, with a
// summoning-sick 3/3 and an open board, is an attack the bot would not
// otherwise have. The same grant with every creature already able to
// attack adds nothing, and a second one while the first is on the stack
// adds nothing either.
func TestTeamHasteLetsFreshCreaturesAttack(t *testing.T) {
	src := cardID(1)
	goro := purphorosGod(src)
	goro.Name = "Goro-Goro"
	goro.ActivatedAbilities[0].Label = "{R}: Creatures you control gain haste until end of turn."
	goro.ActivatedAbilities[0].ManaCost = "{R}"
	goro.ActivatedAbilities[0].Purpose.TeamPump = &protocol.PumpView{Keywords: []string{"haste"}}
	grant := activateMove(t, 0, src, "Goro-Goro: creatures you control gain haste", &legal.MoveCost{Mana: "{R}"})
	fresh := creature(cardID(3), 0, "Fresh Ogre", 3, 3, sick())
	board := append(manaLands(2, 0, 90), goro, fresh)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grant); i != 1 {
		t.Errorf("picked %d (%s), want haste for the fresh 3/3", i, why)
	}
	ready := creature(cardID(3), 0, "Old Ogre", 3, 3)
	v = newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(append(manaLands(2, 0, 90), goro, ready)...), withTurn(5, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grant); i != 0 {
		t.Errorf("the bot gave haste to a creature that can already attack: %s", why)
	}
	pending := protocol.StackItemView{ID: "item-goro", Kind: "ability", Controller: seatID(0).String(), SourceCardID: src}
	v = newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, "precombat_main"), withStackItems(pending))
	moves := []legal.Move{passMove(0), grant}
	if got := rankValue(t, heuristic.New(), input(0, v, moves...), grant.Label); got >= 0 {
		t.Errorf("a second haste grant over the first on the stack is priced %g, want below passing", got)
	}
}

// TestShadowspearWaitsForSomethingToStrip: Shadowspear's "{1}: Permanents
// your opponents control lose hexproof and indestructible until end of
// turn." declares only `restrict`, which #4318 holds until the stack has
// something to answer: in the bot's main phase with no opposing hexproof
// or indestructible permanent, and in its own combat with attackers, it
// passes.
func TestShadowspearWaitsForSomethingToStrip(t *testing.T) {
	src := cardID(1)
	spear := protocol.CardView{
		InstanceID: src, Name: "Shadowspear", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Legendary Artifact — Equipment", ManaCost: "{1}", KnownByYou: true,
	}
	restrict := []string{"restrict"}
	spear.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{1}: Permanents your opponents control lose hexproof and indestructible until end of turn", ManaCost: "{1}",
		Purpose: &protocol.PurposeView{Answers: &restrict},
	}}
	strip := activateMove(t, 0, src, "Shadowspear: strip hexproof and indestructible", &legal.MoveCost{Mana: "{1}"})
	for _, step := range []string{"precombat_main", "declare_attackers", "postcombat_main"} {
		t.Run(step, func(t *testing.T) {
			board := append(manaLands(4, 0, 90), spear, creature(cardID(4), 1, "Their Bear", 2, 2))
			if step == "declare_attackers" {
				board = append(board, creature(cardID(3), 0, "Attacker", 3, 3, attacking(1), tapped()))
			}
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(5, 0, step))
			if i, why := decideOn(t, heuristic.New(), v, passMove(0), strip); i != 0 {
				t.Errorf("the bot paid for Shadowspear with nothing to strip: %s", why)
			}
		})
	}
}

// TestTeamPumpOnBlocksPaysOnlyForAnExchangeItChanges: an opponent's turn,
// the bot's creature blocking before damage, all its mana spare. A 2/2
// blocking a 3/3 kills it with +1/+0, so the bot pumps; a 1/1 chump
// blocking a 6/6 neither kills it nor survives, so the pump buys nothing.
func TestTeamPumpOnBlocksPaysOnlyForAnExchangeItChanges(t *testing.T) {
	src, theirs, mine := cardID(1), cardID(4), cardID(3)
	cases := []struct {
		name       string
		atkP, atkT int
		blkP, blkT int
		wantPump   bool
	}{
		{"a 2/2 blocking a 3/3", 3, 3, 2, 2, true},
		{"a 1/1 chumping a 6/6", 6, 6, 1, 1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			board := teamPumpBoard(src,
				creature(theirs, 1, "Their Attacker", c.atkP, c.atkT, attacking(0), tapped()),
				creature(mine, 0, "Blocker", c.blkP, c.blkT, blocking(theirs)))
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(6, 1, "declare_blockers"))
			i, why := decideOn(t, heuristic.New(), v, passMove(0), purphorosPump(t, src))
			if got := i == 1; got != c.wantPump {
				t.Errorf("pumped=%v (%s), want %v", got, why, c.wantPump)
			}
		})
	}
}
