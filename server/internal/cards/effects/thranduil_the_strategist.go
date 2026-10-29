package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thranduil, the Strategist — Legendary Creature — Elf Noble {3}{G}{U}, 4/4:
//
//	"Other Elves you control have "{T}: Add {G} or {U}."
//	 Landfall — Whenever a land you control enters, create a 1/1 green
//	 Elf creature token."
//
// Tyvar Kell's ability grant (ADR 0093) with "other": every Elf
// permanent the controller controls except Thranduil has the mana
// ability, and a creature granted it is still subject to summoning
// sickness (CR 302.6). The tokens the landfall makes are Elves, so they
// get the grant as they enter. The mana ability is "{G} or {U}", one
// mana chosen as it is activated, which is the "{G|U}" pipe.
//
// No simplification.
func init() {
	const grant = "thranduil-the-strategist/tap-for-g-or-u"
	Register(Spec{
		OracleID:     "7bef91cd-df65-4135-86a6-d17b71470452",
		Name:         "Thranduil, the Strategist",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{TapForManaGrant(grant, "{G|U}", "Add {G} or {U}", "{T}: Add {G} or {U}.")},
		Static: []game.StaticAbility{
			GrantAbilities(thranduilOtherElvesYouControl, grant),
		},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Thranduil, the Strategist — create a 1/1 green Elf", game.LandfallPayoff{Tokens: 1},
				b34CreateTokens(thranduilElfToken, 1)),
		},
	})
}

// thranduilOtherElvesYouControl is "other Elves you control": every Elf
// permanent but the source. "Elf" is a word of the source's text
// (CR 612.2, #3984).
func thranduilOtherElvesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID &&
		target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Elf"))
}

func thranduilElfToken() game.Card { return TokenCard("1/1 green Elf") }
