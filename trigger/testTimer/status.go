package testTimer

func getStatus(t *Trigger, settings *HandlerSettings, stat int) int {
	for {

		select {
		case stat = <-t.ch:
			if stat == Stop {
				settings.Count = 0
				continue
			}
			if stat == Start {
				break
			}
			if stat == Pause {
				t.ch <- Resume
				stat = Resume
			}

		case stat = <-settings.Ch:
			if stat == Start {
				break
			}
			if stat == Stop {
				settings.Count = 0
				continue
			}
			if stat == Pause {
				settings.Ch <- Resume
				stat = Resume
			}
		default:
			if stat == Start || stat == Resume {
				break
			} else {
				continue
			}

		}
		return stat
	}
}
