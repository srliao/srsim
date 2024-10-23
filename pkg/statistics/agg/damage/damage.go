package damage

import (
	calc "github.com/aclements/go-moremath/stats"

	"github.com/simimpact/srsim/pkg/model"
	"github.com/simimpact/srsim/pkg/statistics/agg"
)

func init() {
	agg.Register(NewAgg)
}

type buffer struct {
	iters           uint32
	dmgPerCharPerAV map[int32]map[int32]calc.StreamStats // map of char target id -> av -> damage
}

func NewAgg(cfg *model.SimConfig) (agg.Aggregator, error) {
	out := buffer{
		iters:           cfg.Settings.Iterations,
		dmgPerCharPerAV: make(map[int32]map[int32]calc.StreamStats),
	}
	return &out, nil
}

func (b *buffer) Add(result *model.IterationResult) {
	for charID, d := range result.CharacterDetails {
		if _, ok := b.dmgPerCharPerAV[charID]; !ok {
			b.dmgPerCharPerAV[charID] = make(map[int32]calc.StreamStats)
		}
		for av, val := range d.DamageDealtByAv {
			stat := b.dmgPerCharPerAV[charID][av]
			stat.Add(val)
			b.dmgPerCharPerAV[charID][av] = stat
		}
	}
}

func (b *buffer) Flush(result *model.Statistics) {
	if result.CharacterStatistics == nil {
		result.CharacterStatistics = make(map[int32]*model.CharacterStatistics)
	}
	for charID, data := range b.dmgPerCharPerAV {
		if _, ok := result.CharacterStatistics[charID]; !ok {
			//nolint:exhaustruct // only need empty pointer struct
			result.CharacterStatistics[charID] = &model.CharacterStatistics{}
		}
		if result.CharacterStatistics[charID].DamageDealtByAv == nil {
			result.CharacterStatistics[charID].DamageDealtByAv = make(map[int32]*model.DescriptiveStats)
		}
		for av, val := range data {
			result.CharacterStatistics[charID].DamageDealtByAv[av] = agg.ToDescriptiveStats(&val)
		}
	}
}
