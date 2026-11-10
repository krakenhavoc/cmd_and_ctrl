package game

import "github.com/google/uuid"

// secret_choice.go — "secretly choose <A>, <B>, or <C>" and "Reveal the
// <thing> you chose" (A Killer Among Us, #4469, CR 701.20 and the printed
// ruling). The answer is a word stamped on the permanent
// (Card.SecretChoice) with the player who chose it (Card.SecretChoiceBy).
// It is never projected into a view, and the answer emits no event, so no
// opponent learns it. The reveal (AbilityCost.RevealSecretChoice) is paid
// at activation like any other reveal cost (CR 602.2b), shows the word to
// the table (EventCreatureTypeChosen, the line the log already writes for
// a chosen creature type), and can be paid only by the player who chose:
// a Case that changed hands can't be revealed by its new controller.

// QueueSecretChoiceForEffect asks `chooser` to secretly choose one of
// `options` for the permanent `source` as an effect resolves, and returns
// the choice ID, or uuid.Nil when nothing was queued (no options, or a
// chooser who has left). Caller must hold g.mu.
func (g *Game) QueueSecretChoiceForEffect(chooser, source uuid.UUID, question string, options []string) uuid.UUID {
	if len(options) == 0 {
		return uuid.Nil
	}
	words := append([]string(nil), options...)
	choices := make([]ChoiceOption, len(words))
	for i, w := range words {
		choices[i] = ChoiceOption{Label: w}
	}
	return g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		Options:  choices,
		ThenKey:  secretChoiceThen,
		Carry:    []uuid.UUID{source, chooser},
		Words:    words,
	})
}

// secretChoiceThen stamps the answer on the permanent without announcing
// it. An on-disk key: never renamed, never reused.
var secretChoiceThen OptionPickThen

func init() {
	secretChoiceThen = RegisterOptionPickThen("option-pick/secret-choice", secretChoiceAnswered)
}

func secretChoiceAnswered(g *Game, r OptionPicked) error {
	if len(r.Carry) < 2 || r.Index < 0 || r.Index >= len(r.Words) {
		return nil
	}
	// The permanent is located live: a Case that has left by the time the
	// player answers has nowhere to keep the choice (CR 608.2).
	if i := findCardOnBattlefield(g, r.Carry[0]); i >= 0 {
		g.Battlefield.Cards[i].SecretChoice = r.Words[r.Index]
		g.Battlefield.Cards[i].SecretChoiceBy = r.Carry[1]
	}
	return nil
}

// SecretChoiceOf is the word secretly chosen for the battlefield
// permanent `sourceID`, and by whom; "" when none was chosen.
//
// Caller must hold either lock.
func (g *Game) SecretChoiceOf(sourceID uuid.UUID) (string, uuid.UUID) {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		c := &g.Battlefield.Cards[i]
		return c.SecretChoice, c.SecretChoiceBy
	}
	return "", uuid.Nil
}

// CanRevealSecretChoiceForEffect reports whether `player` may pay the
// reveal cost for `sourceID`: they are the one who chose, and a choice was
// made. Caller must hold g.mu.
func (g *Game) CanRevealSecretChoiceForEffect(player, sourceID uuid.UUID) bool {
	word, by := g.SecretChoiceOf(sourceID)
	return word != "" && by == player
}

// validateRevealSecretChoiceCostLocked refuses a reveal nobody may make.
func (g *Game) validateRevealSecretChoiceCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost) error {
	if !cost.RevealSecretChoice {
		return nil
	}
	if !g.CanRevealSecretChoiceForEffect(playerID, sourceID) {
		return ErrInvalidParam
	}
	return nil
}

// payRevealSecretChoiceCostLocked shows the chosen word to the table. Paid
// before the sacrifices, while the permanent is still on the battlefield.
func (g *Game) payRevealSecretChoiceCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost) {
	if !cost.RevealSecretChoice {
		return
	}
	word, _ := g.SecretChoiceOf(sourceID)
	if word == "" {
		return
	}
	g.EmitEvent(Event{Kind: EventCreatureTypeChosen, Actor: playerID, CardID: sourceID, Label: word})
}
