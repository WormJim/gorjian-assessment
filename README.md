# Gorgian Interview Assessment

This is an implementation of the [IBlaster](https://github.com/WormJim/gorjian-assessment/blob/dc01b829db8b3b048a2559ae9d88f9b7c774bc61/coding_interview%5B37%5D.go#L82C8-L82C9) Interface. It constructs the Queue and Process methods.

The Queue method uses the [IRepo](https://github.com/WormJim/gorjian-assessment/blob/fcc2497450a5eabcd52b04166a9088f0d8bf0e8c/coding_interview%5B37%5D.go#L60) interface on the newly constructed [Blaster struct](https://github.com/WormJim/gorjian-assessment/blob/dc01b829db8b3b048a2559ae9d88f9b7c774bc61/coding_interview%5B37%5D.go#L120) in order to query all the Associates and their eligible contacts. Once fetched, the list of contacts is filtered to remove those who have received an email within the last 7 days. Once cleared, a task is constructed and added to the [IWorker](https://github.com/WormJim/gorjian-assessment/blob/dc01b829db8b3b048a2559ae9d88f9b7c774bc61/coding_interview%5B37%5D.go#L76) queue.

The Process method validates each contact prior to calling the send method on the [Mailer](https://github.com/WormJim/gorjian-assessment/blob/dc01b829db8b3b048a2559ae9d88f9b7c774bc61/coding_interview%5B37%5D.go#L55). It simply ensures each contact and associate exist prior to initializing the mailer.
