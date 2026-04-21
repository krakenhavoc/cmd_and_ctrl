package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path to Exile — "Exile target creature. Its controller may
// search their library for a basic land card, put it onto the
// battlefield tapped, then shuffle."
//
// S14 sandbox simplifications:
//   - "May" is treated as always. If the exiled creature's
//     controller has a basic land, it always fetches.
//   - The fetched land enters UNTAPPED in S14. Enters-tapped is a
//     replacement-effect concern that lands in S17; the "tapped"
//     half of Path's text is deferred alongside Cultivate's.
//   - Capture the controller BEFORE ExileTarget fires — post-move
//     the card is in exile and its controller is still stamped,
//     but the lookup is cleaner to do upfront.
func init() {
	Register(Spec{
		OracleID:   "d683d985-9888-4d21-8b5f-69e69ce4a03b",
		Name:       "Path to Exile",
		TargetMode: "creature",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			targetID := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(targetID)
			if !ok {
				return nil
			}
			controller := card.Controller
			if err := (ExileTarget{Target: targetID}).Apply(ctx); err != nil {
				return err
			}
			// Search the controller's library for any basic land.
			// Predicate matches Scryfall's "Basic Land" prefix on the
			// type line, case-insensitive.
			return SearchLibrary{
				Player: controller,
				Predicate: func(c game.Card) bool {
					// Case-insensitive "Basic Land" substring check.
					lower := []byte(c.TypeLine)
					for i := range lower {
						if lower[i] >= 'A' && lower[i] <= 'Z' {
							lower[i] += 'a' - 'A'
						}
					}
					needle := "basic land"
					if len(lower) < len(needle) {
						return false
					}
					for i := 0; i+len(needle) <= len(lower); i++ {
						match := true
						for j := 0; j < len(needle); j++ {
							if lower[i+j] != needle[j] {
								match = false
								break
							}
						}
						if match {
							return true
						}
					}
					return false
				},
				Dest:    game.ZoneBattlefield,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
			}.Apply(ctx)
		},
	})
}
