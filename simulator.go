package main

import (
	"strconv"
)

const timeIncrement float64 = 0.01

type simulator struct {
	Players []*playerCharacter
	Enemies []*enemy
}

type simResult struct {
	skillCount      int
	critCount       int
	directCount     int
	critDirectCount int
	totalDamageDone int
	timeEllapsed    float64
	skillsPerformed []int
	autoAttacks     int
	dotTicks        int
}

func (s *simulator) RunSkillQueue(SkillQueue []string) simResult {
	currentSkillIndex := 0
	var totalResult simResult
	totalResult.skillCount = 0
	totalResult.critCount = 0
	totalResult.directCount = 0
	totalResult.critDirectCount = 0
	totalResult.totalDamageDone = 0
	totalResult.timeEllapsed = 0.0
	totalResult.skillsPerformed = make([]int, len(s.Players[0].Skills))

	for currentSkillIndex < len(SkillQueue) {
		for i := range s.Players {
			results := enemies[0].advanceTime(timeIncrement)
			if results != nil {
				totalResult.dotTicks += len(results)
			}

			// First advance time to take care of Dots, Cooldowns, and Autos
			timeResult := s.Players[i].advanceTime(timeIncrement, enemies[0])

			if timeResult != nil {
				totalResult.autoAttacks++
				if results != nil {
					results = append(results, *timeResult)
				} else {
					results = []skillResult{*timeResult}
				}
			}
			for i, result := range results {
				enemies[0].applyDamage(&result)
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

			// Next attempt to perform the next action in the queue
			actionResult := s.Players[i].performAction(SkillQueue[currentSkillIndex], enemies[0])
			if actionResult != nil {
				enemies[0].applyDamage(actionResult)
				actionResult.Log()
				actionResult.ApplyBuffs(*s.Players[i])
				totalResult.totalDamageDone += actionResult.DamageDone
				totalResult.skillsPerformed[actionResult.SkillIndex]++
				currentSkillIndex++
				totalResult.skillCount++
				if actionResult.DidCrit {
					totalResult.critCount++
				}
				if actionResult.DidDirect {
					totalResult.directCount++
				}
				if actionResult.DidCrit && actionResult.DidDirect {
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
	globalLogFloat(Important, "Auto Attacks: ", float64(s.autoAttacks))
	globalLogFloat(Important, "DoT Ticks: ", float64(s.dotTicks))
	globalLog(Important, "")
}
