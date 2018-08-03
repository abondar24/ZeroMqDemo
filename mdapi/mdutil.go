package mdapi

func unwrap(msg []string) (head string, tail []string) {
	head = msg[0]
	if len(msg) > 1 && msg[1] == "" {
		tail = msg[2:]
	} else {
		tail = msg[1:]
	}
	return
}

func popStr(ss []string) (s string, ss2 []string) {
	s = ss[0]
	ss2 = ss[1:]
	return
}

func popMsg(msgs [][]string) (msg []string, msgs2 [][]string) {
	msg = msgs[0]
	msgs2 = msgs[1:]
	return
}

func popWorker(workers []*Worker) (worker *Worker, workers2 []*Worker) {
	worker = workers[0]
	workers2 = workers[1:]
	return
}

func delWorker(workers []*Worker, worker *Worker) []*Worker {
	for i := 0; i < len(workers); i++ {
		if workers[i] == worker {
			workers = append(workers[:i], workers[i+1:]...)
			i--
		}
	}
	return workers
}
