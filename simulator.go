package main

import (
	"strconv"
)

const timeIncrement float64 = 0.01

type simulator struct {
	Players    []*playerCharacter
	Enemies    []*enemy
	SkillQueue []string
}

type simResult struct {
	skillCount      int
	critCount       int
	directCount     int
	critDirectCount int
	totalDamageDone int
	timeEllapsed    float64
	skillsPerformed []int
}

func (s *simulator) Run() simResult {
	currentSkillIndex := 0
	var totalResult simResult
	totalResult.skillCount = 0
	totalResult.critCount = 0
	totalResult.directCount = 0
	totalResult.critDirectCount = 0
	totalResult.totalDamageDone = 0
	totalResult.timeEllapsed = 0.0
	totalResult.skillsPerformed = make([]int, len(s.Players[0].Skills))

	for currentSkillIndex < len(s.SkillQueue) {
		for i := range s.Players {
			enemies[0].advanceTime(timeIncrement)
			result := s.Players[i].advanceTime(timeIncrement, enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalResult.totalDamageDone += result.DamageDone
				if result.DidCrit {
					totalResult.critCount++
				}
				if result.DidDirect {
					totalResult.directCount++
				}
				if result.DidCrit && result.DidDirect {
					totalResult.critDirectCount++
				}
			}
			totalResult.timeEllapsed += timeIncrement
			result = s.Players[i].performAction(s.SkillQueue[currentSkillIndex], enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalResult.totalDamageDone += result.DamageDone
				totalResult.skillsPerformed[result.SkillIndex]++
				currentSkillIndex++
				totalResult.skillCount++
				if result.DidCrit {
					totalResult.critCount++
				}
				if result.DidDirect {
					totalResult.directCount++
				}
				if result.DidCrit && result.DidDirect {
					totalResult.critDirectCount++
				}
			}
		}
	}

	return totalResult
}

func (s *simResult) log() {
	globalLog(Important, "Total Damage Done: "+strconv.Itoa(s.totalDamageDone))
	globalLogFloat(Important, "Time ellapsed: ", s.timeEllapsed)
	globalLogFloat(Important, "DPS: ", float64(s.totalDamageDone)/s.timeEllapsed)
	globalLog(Important, "Skills performed: "+strconv.Itoa(s.skillCount))
	globalLogFloat(Important, "Crit Rate: ", float64(s.critCount)/float64(s.skillCount))
	globalLogFloat(Important, "Direct Rate: ", float64(s.directCount)/float64(s.skillCount))
	globalLogFloat(Important, "CritDirect Rate: ", float64(s.critDirectCount)/float64(s.skillCount))
	globalLogIntSlice(Important, "Skill Counts: ", s.skillsPerformed)
	globalLog(Important, "")
}
