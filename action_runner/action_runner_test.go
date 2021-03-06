package action_runner_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/executor/action_runner"
)

type FakeAction struct {
	perform func(result chan<- error)
	cancel  func()
	cleanup func()
}

func (fakeAction FakeAction) Perform(result chan<- error) {
	if fakeAction.perform != nil {
		fakeAction.perform(result)
	} else {
		result <- nil
	}
}

func (fakeAction FakeAction) Cancel() {
	if fakeAction.cancel != nil {
		fakeAction.cancel()
	}
}

func (fakeAction FakeAction) Cleanup() {
	if fakeAction.cleanup != nil {
		fakeAction.cleanup()
	}
}

var _ = Describe("ActionRunner", func() {
	It("performs them all in order and sends back nil", func(done Done) {
		defer close(done)

		seq := make(chan int, 3)

		runner := New([]Action{
			FakeAction{
				perform: func(result chan<- error) {
					seq <- 1
					result <- nil
				},
			},
			FakeAction{
				perform: func(result chan<- error) {
					seq <- 2
					result <- nil
				},
			},
			FakeAction{
				perform: func(result chan<- error) {
					seq <- 3
					result <- nil
				},
			},
		})

		result := make(chan error)
		go runner.Perform(result)

		Ω(<-seq).Should(Equal(1))
		Ω(<-seq).Should(Equal(2))
		Ω(<-seq).Should(Equal(3))

		Ω(<-result).Should(BeNil())
	})

	It("cleans up the actions in reverse order before sending back the result", func(done Done) {
		defer close(done)

		cleanup := make(chan int, 3)

		runner := New([]Action{
			FakeAction{
				cleanup: func() {
					cleanup <- 1
				},
			},
			FakeAction{
				cleanup: func() {
					cleanup <- 2
				},
			},
			FakeAction{
				cleanup: func() {
					cleanup <- 3
				},
			},
		})

		result := make(chan error)
		go runner.Perform(result)

		Ω(<-cleanup).Should(Equal(3))
		Ω(<-cleanup).Should(Equal(2))
		Ω(<-cleanup).Should(Equal(1))

		Ω(<-result).Should(BeNil())
	})

	Context("when an action fails in the middle", func() {
		It("sends back the error and does not continue performing, and cleans up completed actions", func(done Done) {
			defer close(done)

			disaster := errors.New("oh no!")

			seq := make(chan int, 3)
			cleanup := make(chan int, 3)

			runner := New([]Action{
				FakeAction{
					perform: func(result chan<- error) {
						seq <- 1
						result <- nil
					},
					cleanup: func() {
						cleanup <- 1
					},
				},
				FakeAction{
					perform: func(result chan<- error) {
						result <- disaster
					},
					cleanup: func() {
						cleanup <- 2
					},
				},
				FakeAction{
					perform: func(result chan<- error) {
						seq <- 3
						result <- nil
					},
					cleanup: func() {
						cleanup <- 3
					},
				},
			})

			result := make(chan error)
			go runner.Perform(result)

			Ω(<-seq).Should(Equal(1))
			Ω(<-cleanup).Should(Equal(1))

			Ω(<-result).Should(Equal(disaster))

			Consistently(seq).ShouldNot(Receive())
			Consistently(cleanup).ShouldNot(Receive())
		})
	})

	Context("when the runner is canceled in the middle", func() {
		It("cancels the running action and waits for completed actions to be cleaned up", func(done Done) {
			defer close(done)

			seq := make(chan int, 3)
			cleanup := make(chan int, 3)

			waitingForInterrupt := make(chan bool)
			interrupt := make(chan bool)
			interrupted := make(chan bool)

			startCleanup := make(chan bool)

			runner := New([]Action{
				FakeAction{
					perform: func(result chan<- error) {
						seq <- 1
						result <- nil
					},
					cleanup: func() {
						<-startCleanup
						cleanup <- 1
					},
				},
				FakeAction{
					perform: func(result chan<- error) {
						seq <- 2

						waitingForInterrupt <- true
						<-interrupt
						interrupted <- true

						result <- nil
					},
					cancel: func() {
						interrupt <- true
					},
					cleanup: func() {
						cleanup <- 2
					},
				},
				FakeAction{
					perform: func(result chan<- error) {
						seq <- 3
						result <- nil
					},
					cleanup: func() {
						cleanup <- 3
					},
				},
			})

			result := make(chan error)
			go runner.Perform(result)

			Ω(<-seq).Should(Equal(1))
			Ω(<-seq).Should(Equal(2))

			<-waitingForInterrupt

			cancelled := runner.Cancel()

			<-interrupted

			Consistently(cancelled).ShouldNot(Receive())

			startCleanup <- true

			<-cancelled

			Ω(<-cleanup).Should(Equal(1))

			Ω(<-result).Should(Equal(CancelledError))

			Consistently(seq).ShouldNot(Receive())
			Consistently(cleanup).ShouldNot(Receive())
		})
	})
})
