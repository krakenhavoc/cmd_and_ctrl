package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// King of the Pride — 2/1 Creature — Cat for {2}{W} (EDHREC rank
// 4013):
//
//	"Other Cats you control get +2/+1."
//
// The Cat lord with the biggest number on it, which is why every
// Arahbo and Mirri list runs it over the +1/+1 versions. In the batch
// as the "other … you control" corner of the lord builder: the
// filter has to exclude the King himself AND everybody else's Cats,
// and getting either half wrong is invisible on a board of one.
//
// Layer 7c through TribalAnthem, so it reads the POST-layer types: a
// changeling is a Cat and gets the bonus, and a Cat that something
// turned into a Wall stops getting it. Two Kings pump each other and
// neither pumps itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96d0e4dd-6cc2-4349-ac44-785b50f8dd90",
		Name:         "King of the Pride",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Cat"}, Others: true, YoursOnly: true}, 2, 1),
		},
	})
}
