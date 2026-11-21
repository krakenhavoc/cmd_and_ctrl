package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_shields_test.go — ADR 0106 PR 2 (#1806): the printed
// "<these> spells can't be countered" statics, end to end through the
// real catalog and the engine's counter gate. The gate's own contract
// (every verb, every form, the source leaving) is pinned with stubs in
// game/cant_be_countered_test.go.

const (
	CsGaeasHeraldOracle         = "f21c3d22-6149-475f-ae6e-37a25ef020e5"
	CsLeylineOfLifeforceOracle  = "997478aa-b790-4269-a626-abf0cb30fea0"
	CsProwlingSerpopardOracle   = "4bdfd718-f474-4223-88d8-7fa9fb0c86b4"
	CsRootSliverOracle          = "cb752313-dc7f-47fc-9077-9e3298e8f5fb"
	CsSphinxFinalWordOracle     = "d4246e4d-390d-4925-a5a8-89cd096a237c"
	CsSpellbreakerOracle        = "cba07472-7212-4411-a9f9-38a48870ad69"
	CsSurrakOracle              = "7c766365-82c2-46d6-8521-42e73129f4ef"
	CsAllosaurusShepherdOracle  = "98295bdc-db58-4e1f-ad4c-63092aea207e"
	CsHexingSquelcherOracle     = "96c2d71e-0fc9-42aa-bc6b-6d0ae3f66f8b"
	CsDestinySpinnerOracle      = "7dd189f8-a23f-4b43-b0ef-fb174a9664ba"
	CsThryxOracle               = "20004398-bd04-4583-8378-9df743246a32"
	CsCunningNightbonderOracle  = "7351317d-62a1-4894-9811-aa950c49eece"
	CsLierOracle                = "f224db3d-cdd2-43f2-b625-4b861efa6449"
	CsTestGraveyardSorceryOracl = "test-1806-graveyard-sorcery"
)

// csSpell is a spell to put on the stack: its card, who cast it, who
// controls it now, and whether it is a copy.
type CsSpell struct {
	Card               game.Card
	Caster, Controller int // seat index; seat 0 controls the static
	Copy               bool
}

func CsCard(name, typeLine, cost string, power int, keywords ...string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Keywords: keywords}
}

var (
	CsBlueInstant  = CsCard("Opt", "Instant", "{U}", 0)
	CsGreenInstant = CsCard("Giant Growth", "Instant", "{G}", 0)
	CsSorcery      = CsCard("Divination", "Sorcery", "{2}{U}", 0)
	CsBear         = CsCard("Grizzly Bears", "Creature — Bear", "{1}{G}", 2)
	CsBigCreature  = CsCard("Craw Wurm", "Creature — Wurm", "{4}{G}{G}", 6)
	CsEnchantment  = CsCard("Pacifism", "Enchantment — Aura", "{1}{W}", 0)
	CsSliver       = CsCard("Muscle Sliver", "Creature — Sliver", "{1}{G}", 1)
	CsChangeling   = CsCard("Changeling Outcast", "Creature — Shapeshifter", "{B}", 1, "changeling")
	CsFiveDrop     = CsCard("Mind Spring", "Sorcery", "{4}{U}", 0)
	CsFourDrop     = CsCard("Concentrate", "Sorcery", "{2}{U}{U}", 0)
	CsFlashBeast   = CsCard("Ambush Viper", "Creature — Snake", "{1}{G}", 2, "flash")
)

// csPushSpell puts a spell on the stack the way a cast leaves it: the
// card on g.Stack and its item in StackMeta.
func CsPushSpell(g *game.Game, s CsSpell) uuid.UUID {
	id := uuid.New()
	caster, controller := g.Seats[s.Caster], g.Seats[s.Controller]
	c := s.Card
	c.InstanceID, c.Owner, c.Controller = id, caster.ID, controller.ID
	g.WithWriteLock(func() {
		g.Stack.PushTop(c)
		item := &game.StackItem{ID: id, Kind: game.StackItemSpell, Controller: controller.ID,
			Owner: caster.ID, SourceCardID: id, IsCopy: s.Copy}
		if caster.ID != controller.ID {
			item.BaseController = caster.ID
		}
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[id] = item
	})
	return id
}
