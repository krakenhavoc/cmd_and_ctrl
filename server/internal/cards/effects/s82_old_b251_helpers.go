package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// EntersWithOneMore is the count of a cast trigger's "that planeswalker
// enters with an additional loyalty counter on it" (Dark Intimations): one
// counter whatever was paid, for WheneverYouCastThatEntersWithCounters.
func EntersWithOneMore(game.ManaSpent) int { return 1 }
