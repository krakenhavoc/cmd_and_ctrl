package effectstest

// multi_keyword_cards_test.go covers the S18 sub-PR 5 batch 2
// catalog cards. Each card carries ≥2 keywords; the tests assert
// every keyword lands in Effective().Abilities on the battlefield,
// AND that off-battlefield HasKeyword returns true via the
// catalog fallback (the forcing function for flash gating on
// hand-resident Ambush Viper).

const (
	VampireNighthawkOracle = "feb244f8-bcb1-44cf-9940-2719221a7309"
	BaneslayerAngelOracle  = "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe"
	AmbushViperOracle      = "8957f7c2-040c-4048-9f21-efa7c97682b7"
	BoggartBruteOracle     = "880793a2-84b0-4ae1-9bb8-6e942e3dc723"
)
