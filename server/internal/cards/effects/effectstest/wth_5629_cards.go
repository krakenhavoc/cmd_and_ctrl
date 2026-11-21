package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wth_5629_cards_test.go — a batch of Weatherlight cards (#5629). Every
// test plays the real catalog entry through the engine.

const (
	WthDuskriderFalconOracle        = "d87ab916-5d49-4840-8d5b-805af800a83c"
	WthRazortoothRatsOracle         = "72e9a207-dad9-4c51-8040-e90a36bd35e6"
	WthBenalishKnightOracle         = "32b401e9-163f-4917-a728-fc63b25ef602"
	WthCloudDjinnOracle             = "4914d38d-ed6b-4531-b72c-06f0968c44e9"
	WthBloodrockCyclopsOracle       = "06a3fe24-7e1f-4f98-ba2f-dbb5d38f4b9d"
	WthForiysianBrigadeOracle       = "7261dc6d-7669-4682-977b-1304f57a201d"
	WthTolarianDrakeOracle          = "d999f2a3-5cb4-4f12-a79a-c5db0a300620"
	WthShadowRiderOracle            = "a518b31a-7e9a-460d-839f-39cb1de4f570"
	WthSerrasBlessingOracle         = "49cdd05c-feeb-4c24-9d88-069f5d2e08c3"
	WthUktabiEfreetOracle           = "2fd9a37f-151a-4501-9d8e-bc67cc902592"
	WthGallowbraidOracle            = "4a4418fa-7574-4027-aa09-4ca789be15ba"
	WthMorinfenOracle               = "d69f3aaf-8462-432d-afaf-4176ac04962b"
	WthArcticWolvesOracle           = "c9e7218f-07f2-46f3-bc63-eb5c0d32538b"
	WthFledglingDjinnOracle         = "2c73ef77-ae58-401f-8747-538c4cd075d0"
	WthStripedBearsOracle           = "ac856e91-1b88-464b-9420-dc311ce814ea"
	WthMerfolkTradersOracle         = "8af6646a-10ba-448b-bd5e-48feb70de5e5"
	WthLavaHoundsOracle             = "86c0dc85-d146-4dfa-819e-f78835bfffff"
	WthSageOwlOracle                = "d47fc902-51f7-4ad1-8ee4-c973e32192b8"
	WthOdylicWraithOracle           = "613fa223-856e-49c4-8952-224895e00b5e"
	WthBogardanFirefiendOracle      = "1d61f3a7-5861-479b-8312-f2d308c35b68"
	WthAbyssalGatekeeperOracle      = "4954181c-cb46-4cac-adc3-2c8b7693a211"
	WthHurloonShamanOracle          = "9fc74fc9-bb02-4d9e-bb8e-f44f219d8946"
	WthTolarianSerpentOracle        = "684e18bf-1390-4d12-8e7a-e8562db64dbb"
	WthMischievousPoltergeistOracle = "c12b5b9a-a46e-4635-8603-62c2ca5ee58e"
	WthHeavyBallistaOracle          = "b4219836-c802-49e2-9a66-f03ca8f13895"
	WthSouthernPaladinOracle        = "cab96aa8-13e3-4dd4-9d44-9e298f807388"
	WthTouchstoneOracle             = "92382bdb-84cf-4456-8ca1-8a19badae7ad"
	WthInfernalTributeOracle        = "a8bf79d2-29fd-4d2f-ac07-2e61d3886bc0"
	WthArgivianFindOracle           = "0b6c74a7-d7d3-4c05-84b1-bcb580bb476f"
	WthRelearnOracle                = "13cb7d87-ff12-4742-8f54-f3175536db40"
	WthArgivianRestorationOracle    = "d887919a-68d5-455d-977c-4f1f0b4b3141"
	WthFitOfRageOracle              = "69d08521-74e1-4215-9ed4-3f12137332a4"
	WthGuidedStrikeOracle           = "a215e813-df7e-4f21-9718-eb264eb62813"
	WthThunderboltOracle            = "6ffae2fa-511a-45bf-93c4-5d53c85324f5"
	WthFatalBlowOracle              = "16774611-c004-4383-b4e1-4a43a9b1f4aa"
	WthTendrilsOfDespairOracle      = "3b00f415-2496-4333-be89-601fd837fe7c"
	WthAbjureOracle                 = "dd97e457-07eb-4918-b780-ae7ff42a37e7"
	WthGerrardsWisdomOracle         = "3e30e8f2-f437-4620-a3bc-dd9c29ae6570"
	WthBlossomingWreathOracle       = "fbd99a88-8926-455e-8e23-58fed6d14719"
	WthNaturesResurgenceOracle      = "276f49ee-7cf3-4a4d-9b14-74aedcbef69f"

	WthBearType    = "Creature — Bear"
	WthRaiderType  = "Creature — Human"
	WthHorrorType  = "Legendary Creature — Phyrexian Horror"
	WthFirstStrike = "first strike"
)

// wthGraveyardCard puts a card of the given type line into p's graveyard.
func WthGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// wthCastWithModes casts a catalog spell from the active player's hand in
// a main phase and returns the announce error.
func WthCastWithModes(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := B13HandCard(active, name, typeLine, oracle)
	AdvanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}
