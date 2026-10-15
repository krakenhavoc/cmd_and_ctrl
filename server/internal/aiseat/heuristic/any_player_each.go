package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// any_player_each.go — ADR 0187 §7 (#3573): the bot as one player of a
// per-object any-player offer (Cleansing: "destroy all lands unless any
// player pays 1 life for each land they choose not to destroy this way").
//
// Every player in APNAP order chooses which objects they will pay life
// for, knowing what the earlier players chose, and then every payment is
// made (CR 101.4). Two players who pay for the same land both pay and the
// land is saved once, so the second payment buys nothing.
//
// The rules the bot follows (owner decision 6):
//   - it pays only for its OWN lands that no earlier player chose, because
//     a land someone else already pays for is saved without it;
//   - it never pays for an opponent's land, whoever controls the effect;
//   - it pays for as many as keep it at or above phyrexianLifeFloor with
//     each land costing the offer's EachLife, the most valuable first.
//
// A later player may still choose the bot's land too. The bot cannot see
// that coming, and paying twice is then that player's choice.
//
// The enumerator offers "pay for none", the suggested set and the
// budgeted combinations; this scores each by those rules, so the highest
// is the answer the rules above describe.

// anyPlayerEachValue scores one answer to a leg of a per-object
// any-player offer. ok is false when the prompt is not one.
func (st *state) anyPlayerEachValue(ch *protocol.PendingChoiceView, named []string, lookup func(string) *protocol.CardView) (float64, string, bool) {
	if ch == nil || ch.AnyPlayerOffer == nil || ch.AnyPlayerOffer.EachLife <= 0 {
		return 0, "", false
	}
	if len(named) == 0 {
		return 0.5, "any player: pay for none", true
	}
	chosen := map[string]bool{}
	for _, e := range ch.EarlierChoices {
		for _, id := range e.CardIDs {
			chosen[id] = true
		}
	}
	var worth float64
	for _, id := range named {
		c := lookup(id)
		switch {
		case c == nil || c.Controller != st.me:
			return 0.25, "any player: never pay for another player's land", true
		case chosen[id]:
			return 0.25, "any player: an earlier player already pays for that land", true
		}
		if v := st.permanentValue(c); v > 0 {
			worth += v
		}
	}
	if st.myLife()-len(named)*ch.AnyPlayerOffer.EachLife < phyrexianLifeFloor {
		return phyrexianLifeDeclined, "any player: the life would take it below the floor", true
	}
	// More lands saved is always better than fewer, and among equally
	// many the more valuable ones; the 0.4 spread keeps every payment
	// above the "pay for none" 0.5 and below a certain win.
	return 0.5 + 0.4*worth/(1+worth), "any player: pay for my lands nobody else pays for", true
}
