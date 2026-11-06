package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wan Shi Tong, All-Knowing — Legendary Creature — Bird Spirit
// {3}{U}{U}, 4/4 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying
//	 When Wan Shi Tong enters, target nonland permanent's owner puts it
//	 into their library second from the top or on the bottom.
//	 Whenever one or more cards are put into a library from anywhere,
//	 create two 1/1 colorless Spirit creature tokens with "This token
//	 can't block or be blocked by non-Spirit creatures.""
//
// The enters trigger is Temporal Cleansing's tuck: the permanent's OWNER
// chooses second from the top or the bottom. The second trigger is one
// per event batch (avB29WheneverCardsArePutIntoALibrary); it does not
// trigger on a scry, a surveil or any other reordering (the ruling), and
// a token put into a library is not a card. The Spirits are Avatar
// Kuruk's printed token, block rule included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6927bd2e-cb05-487a-af86-89ede6e8dabd",
		Name:            "Wan Shi Tong, All-Knowing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Wan Shi Tong, All-Knowing — its owner puts target nonland permanent into their library second from the top or on the bottom",
				avB29WanShiTongTuck),
				TargetPermanent("target nonland permanent", Not(Land()))),
			avB29WheneverCardsArePutIntoALibrary("Wan Shi Tong, All-Knowing — create two 1/1 colorless Spirit creature tokens",
				Do(CreateToken{Template: kurukSpiritToken(), N: 2})),
		},
	})
}

// avB29WanShiTongTuck is the enters trigger's resolution: the target's
// owner puts it second from the top or on the bottom. A target that has
// become illegal does nothing (CR 608.2b).
func avB29WanShiTongTuck(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return PutIntoLibraryAtDepthOrBottom{
			Card:  t.ID,
			Depth: 2,
			Label: "Wan Shi Tong, All-Knowing — put it into your library second from the top or on the bottom",
		}.Apply(ctx)
	}
	return nil
}
