package legal

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A sacrifice cost is an outlet when it can eat a creature, and a price
// when it names a land, a token kind or a plain artifact.
func TestSacrificeInteracts(t *testing.T) {
	outlets := []string{
		"a creature", "another creature", "an artifact or creature", "a permanent",
		"another black creature", "a Goblin", "two Eldrazi", "an artifact creature",
	}
	prices := []string{
		"a land", "a Food", "a Treasure", "three Clues", "two artifacts",
		"a noncreature artifact", "a Swamp and a Forest", "a Desert", "a token",
	}
	for _, l := range outlets {
		if !sacrificeInteracts(&game.TargetSpec{Label: l}) {
			t.Errorf("sacrifice %q should be an outlet", l)
		}
	}
	for _, l := range prices {
		if sacrificeInteracts(&game.TargetSpec{Label: l}) {
			t.Errorf("sacrifice %q is a price, not an outlet", l)
		}
	}
	if sacrificeInteracts(nil) {
		t.Error("no sacrifice cost is not an outlet")
	}
}

// ADR 0142 owner answer 3: with the text read gone, a row that declares
// nothing (an ability carried on a card instance) interacts, and a
// declared row is read from its declaration alone, whatever its label
// says.
func TestUndeclaredRowInteracts(t *testing.T) {
	cases := []struct {
		name      string
		ab        game.ActivatedAbilityShape
		targeted  bool
		interacts bool
		combat    bool
	}{
		{name: "undeclared draw", ab: game.ActivatedAbilityShape{Label: "{T}: Draw a card."}, interacts: true},
		{name: "undeclared, announced with a target", ab: game.ActivatedAbilityShape{Label: "{T}: Draw a card."}, targeted: true},
		{name: "declared value, printed regenerate", ab: game.ActivatedAbilityShape{
			Label: "{1}{G}: Regenerate this creature.", Purpose: game.Purpose{Answers: game.AnswerValue},
		}},
		{name: "declared protect, printed draw", ab: game.ActivatedAbilityShape{
			Label: "{T}: Draw a card.", Purpose: game.Purpose{Answers: game.AnswerProtect},
		}, interacts: true},
		{name: "declared combat grant", ab: game.ActivatedAbilityShape{
			Label: "{B}: This creature gains deathtouch until end of turn.", Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
		}, combat: true},
	}
	for _, c := range cases {
		var targets []game.TargetRef
		if c.targeted {
			targets = []game.TargetRef{{Kind: game.TargetCard}}
		}
		interacts, combat := untargetedFlags(nil, nil, game.ZoneBattlefield, c.ab, targets)
		if interacts != c.interacts || combat.any != c.combat {
			t.Errorf("%s: interacts=%v combat_interacts=%v, want %v/%v", c.name, interacts, combat.any, c.interacts, c.combat)
		}
	}
}

// ADR 0142 amendment (2026-10-10): an "up to one target" row is announced
// with no target beside its targeted moves. The empty announcement reads
// the row's declaration, so a row declared value does not interact, and a
// row declared protect does; a targeted announcement carries has_targets
// and sets neither.
func TestUpToOneTargetRowReadsItsDeclaration(t *testing.T) {
	upTo := &game.TargetSpec{Min: 0, Max: 1}
	value := game.ActivatedAbilityShape{
		Label:   "{2}, Exile this artifact: Shuffle any number of target cards from your graveyard into your library.",
		Targets: upTo, Purpose: game.Purpose{Answers: game.AnswerValue},
	}
	protect := game.ActivatedAbilityShape{
		Label:   "{T}: Choose any number of target creatures. Each gains persist until end of turn.",
		Targets: upTo, Purpose: game.Purpose{Answers: game.AnswerProtect},
	}
	if interacts, combat := untargetedFlags(nil, nil, game.ZoneBattlefield, value, nil); interacts || combat.any {
		t.Errorf("value, no target: interacts=%v combat=%v, want neither", interacts, combat.any)
	}
	if interacts, _ := untargetedFlags(nil, nil, game.ZoneBattlefield, protect, nil); !interacts {
		t.Error("protect, no target: want interacts")
	}
	targeted := []game.TargetRef{{Kind: game.TargetCard}}
	if interacts, combat := untargetedFlags(nil, nil, game.ZoneBattlefield, protect, targeted); interacts || combat.any {
		t.Error("a targeted announcement sets no interacts (has_targets stops you)")
	}
}
