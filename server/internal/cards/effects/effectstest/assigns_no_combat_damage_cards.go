package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assigns_no_combat_damage_cards_test.go — the "attacks and isn't blocked,
// you may <do>. If you do, this creature assigns no combat damage this
// turn" cards of #3531 (CR 510.1a, the gate of ADR 0187 §4).

const (
	NcdOphidian        = "953bdefa-4639-4524-af5c-5055d1289f40"
	NcdFarrelsZealot   = "0d063c0e-66bc-4a9c-98aa-a381b9cea120"
	NcdVigilantes      = "b49baf80-237c-41ee-8a22-2b6a68a313e7"
	NcdKeeper          = "ce976cb2-2b69-4d00-918c-d4900244ac86"
	NcdFloralSpuzzem   = "994de451-14f9-466f-a56e-da052b4666e5"
	NcdKukemssaPirates = "280148c5-3a2f-48f7-a147-4578f2068073"
	NcdSquatters       = "ce936736-915f-44bb-a709-b3bdcf90a3be"
	NcdGoblinVandal    = "50bebbb9-01b7-4eb6-8efd-307c9bcd2517"
	NcdBadger          = "cf9fc317-3184-4999-84ef-945a8842096e"
	NcdBoneDancer      = "13b5c9ef-d0e3-453d-ae5e-8de4f310cc38"
	NcdWhelp           = "b16b10ec-a6a3-4bda-a561-9b7bb8eee194"
	NcdGrunt           = "f973daa7-241e-4eb7-9d98-9ea7557caa3a"
	NcdWarrior         = "bf957e0b-3070-4a85-a63d-3dadd7364d53"
	NcdTitan           = "0d9ac5e0-2b4e-4e7b-937e-9e9bfce7e96b"
	NcdPaladin         = "c4161335-1150-4f0d-bd2b-9211efb3bea6"
	NcdCloak           = "5d258409-28d3-4eac-9f1a-3e494db818bf"
	NcdMantle          = "ffae3ff4-67f7-4e41-b166-c63f88c7aa08"
	NcdRig             = "0b583980-1508-4247-8a28-7127241d8ec0"
)

// ncdPlay attacks seat 1 with `attacker` (nobody blocks), answers every
// prompt the trigger raises (`yes` to the "you may", `targets` to a target
// pick) until the table is quiet, and walks into the combat damage step.
func NcdPlay(t *testing.T, g *game.Game, attacker uuid.UUID, yes bool, targets ...uuid.UUID) {
	t.Helper()
	NcdPlayBlocked(t, g, attacker, uuid.Nil, yes, targets...)
}

// ncdPlayBlocked is ncdPlay with `blocker` (when not uuid.Nil) blocking
// the attacker.
func NcdPlayBlocked(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, yes bool, targets ...uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	Pr7bAttack(t, g, opp.ID, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if blocker != uuid.Nil {
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	FinishBlockDeclarations(t, g)
loop:
	for i := 0; i < 24; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		switch {
		case open == nil && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0:
			break loop
		case open == nil:
			if err := g.PassPriority(); err != nil {
				break loop
			}
		case open.Kind == game.PendingChoicePickTarget:
			var refs []game.TargetRef
			for _, id := range targets {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			if err := g.ResolvePickTargets(open.ID, me.ID, refs); err != nil {
				t.Fatalf("pick targets: %v", err)
			}
		case open.Kind == game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(open.ID, me.ID, yes); err != nil {
				t.Fatalf("confirm: %v", err)
			}
		case open.Kind == game.PendingChoicePayUnless:
			if err := g.ResolvePayUnless(open.ID, me.ID, yes); err != nil {
				t.Fatalf("pay: %v", err)
			}
		default:
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
	}
	Pr7bDamageStep(t, g)
}

// ncdLifeLost is how much life the defending seat lost to combat damage.
func NcdLifeLost(g *game.Game, before int) int { return before - g.Seats[1].Life }

func NcdAttacker(g *game.Game, name, oracle string, p, tough int) uuid.UUID {
	return ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{Name: name, TypeLine: "Creature — Test", OracleID: oracle, Power: p, Toughness: tough})
}

func NcdMarked(g *game.Game, id uuid.UUID) bool {
	return g.PermanentAssignsNoCombatDamageForEffect(id)
}

func NcdAttach(t *testing.T, g *game.Game, aura, host uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
}
