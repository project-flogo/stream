package streamtester

func getStatus(t *Trigger, emitInfo *HandlerEmitterInfo, stat int) int {
	for {

		select {
		case stat = <-emitInfo.Ch:
			if stat == Start {
				break
			}
			if stat == Stop {
				emitInfo.CurentIndex = 0
				continue
			}
			if stat == Pause {
				emitInfo.Ch <- Resume
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
