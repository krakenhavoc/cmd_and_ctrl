package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castFor puts a card in `p`'s hand and casts it from there with the
// given params, advancing to a main phase first. Unlike
// castCatalogSpell the caster need not be the active seat.
func CastFor(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func RtCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// answerRetargetWith answers the open retarget prompt with `answer`
// (nil declines).
func AnswerRetargetWith(t *testing.T, g *game.Game, chooser uuid.UUID, answer []game.TargetRef) {
	t.Helper()
	prompt := LatestRetarget(g, chooser)
	if prompt == nil {
		t.Fatal("no retarget prompt is open")
	}
	if err := g.ResolveRetarget(prompt.ID, chooser, answer); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
}

func RtPlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// boltFrom casts a Lightning Bolt at a player from seat `from`.
func BoltFrom(t *testing.T, g *game.Game, from *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	return CastFor(t, g, from, "Lightning Bolt", "Instant", LightningBoltOracle,
		game.CastSpellParams{Targets: RtPlayerRef(at)})
}

func TwoTargetBiteDown(t *testing.T, g *game.Game, caster *game.Player, mine, theirs uuid.UUID) uuid.UUID {
	t.Helper()
	return CastFor(t, g, caster, "Bite Down", "Instant", B26BiteDownOracle, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}})
}
