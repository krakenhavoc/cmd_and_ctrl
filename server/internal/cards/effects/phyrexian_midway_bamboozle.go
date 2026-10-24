package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Midway Bamboozle — Stickers (an Unfinity sticker sheet, CR
// 123.2; ADR 0208):
//
//	Name stickers "Phyrexian", "Midway" and "Bamboozle"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, you get {TK}.
//	{TK}{TK}{TK} — Undying
//	{TK}{TK} — 4/2
//	{TK}{TK}{TK}{TK}{TK} — 6/9
//
// The first ability sticker gives the object it is on a triggered ability,
// granted as a bundle (ADR 0093): "this creature" is that object and "you"
// its controller. The second gives undying, which the engine runs off the
// dying permanent's last-known abilities (CR 702.93a), so the sticker
// returns the creature and comes back with it (CR 123.5). The name, power
// and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.

// bamboozleAttackTicket is the attack trigger's bundle. On-disk identity:
// never renamed, never reused.
const bamboozleAttackTicket = "phyrexian-midway-bamboozle/attacks-get-ticket"

func init() {
	Register(Spec{
		OracleID:     "dde09abb-e3d3-4c76-b7e8-812949dd67f3",
		Name:         "Phyrexian Midway Bamboozle",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: bamboozleAttackTicket,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Phyrexian Midway Bamboozle sticker — you get {TK}", Do(GetTickets{N: 1})),
			},
			Text: "Whenever this creature attacks, you get {TK}.",
		}},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: bamboozleAttackTicket},
			{Keywords: []string{game.KeywordUndying}},
		}},
	})
}
