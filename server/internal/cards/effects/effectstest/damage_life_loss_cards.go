package effectstest

// damage_life_loss_cards_test.go is #2105 through the cards: damage
// that costs no life (infect, CR 702.90b; a life total that can't
// change, CR 119.8) is not a life loss to any of the three kinds of
// reader — an opponent's loss (Exquisite Blood, b04OpponentLostLife),
// your own loss (Vilis, s22PlayerLostLife), and the turn tally behind
// every "lost life this turn" card (b18LifeLostThisTurn) — while
// ordinary damage still is. Then Lich's Mastery (#2065), which the fix
// unblocked.

const LichsMasteryOracle = "785c306c-471c-4699-8ab2-43c253d569cf"
