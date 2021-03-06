package fakestoreadapter_test

import (
	"errors"
	"github.com/cloudfoundry/storeadapter"
	. "github.com/cloudfoundry/storeadapter/fakestoreadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fakestoreadapter", func() {
	var adapter *FakeStoreAdapter
	var breakfastNode, lunchNode, firstCourseDinnerNode, secondCourseDinnerNode, randomNode storeadapter.StoreNode

	BeforeEach(func() {
		adapter = New()
		breakfastNode = storeadapter.StoreNode{
			Key:   "/menu/breakfast",
			Value: []byte("waffle"),
		}
		lunchNode = storeadapter.StoreNode{
			Key:   "/menu/lunch",
			Value: []byte("burger"),
		}
		firstCourseDinnerNode = storeadapter.StoreNode{
			Key:   "/menu/dinner/first",
			Value: []byte("caesar salad"),
		}
		secondCourseDinnerNode = storeadapter.StoreNode{
			Key:   "/menu/dinner/second",
			Value: []byte("steak"),
		}
		randomNode = storeadapter.StoreNode{
			Key:   "/random",
			Value: []byte("17"),
		}

		err := adapter.SetMulti([]storeadapter.StoreNode{
			breakfastNode,
			lunchNode,
			firstCourseDinnerNode,
			secondCourseDinnerNode,
			randomNode,
		})
		Ω(err).ShouldNot(HaveOccurred())

		adapter.SetErrInjector = NewFakeStoreAdapterErrorInjector("dom$", errors.New("injected set error"))
		adapter.GetErrInjector = NewFakeStoreAdapterErrorInjector("dom$", errors.New("injected get error"))
		adapter.ListErrInjector = NewFakeStoreAdapterErrorInjector("dom$", errors.New("injected list error"))
		adapter.DeleteErrInjector = NewFakeStoreAdapterErrorInjector("dom$", errors.New("injected delete error"))
		adapter.CreateErrInjector = NewFakeStoreAdapterErrorInjector("dom$", errors.New("injected create error"))
	})

	It("should satisfy the interface", func() {
		var adapterInterface storeadapter.StoreAdapter
		adapterInterface = adapter

		Expect(adapterInterface)
	})

	Describe("Creating", func() {
		Context("when creating an existing key", func() {
			It("should error", func() {
				err := adapter.Create(firstCourseDinnerNode)
				Ω(err).Should(Equal(storeadapter.ErrorKeyExists))
			})
		})

		Context("when creating a new key", func() {
			It("should", func() {
				thirdCourseDinnerNode := storeadapter.StoreNode{
					Key:   "/menu/dinner/third",
					Value: []byte("mashed potaters"),
				}

				err := adapter.Create(thirdCourseDinnerNode)
				Ω(err).ShouldNot(HaveOccurred())

				value, err := adapter.Get("/menu/dinner/third")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(thirdCourseDinnerNode))
			})
		})

		Context("when the key matches the error injector", func() {
			It("should return the injected error", func() {
				thirdCourseDinnerNode := storeadapter.StoreNode{
					Key:   "/menu/dinner/random",
					Value: []byte("mashed potaters"),
				}

				err := adapter.Create(thirdCourseDinnerNode)
				Ω(err).Should(Equal(errors.New("injected create error")))
			})
		})
	})

	Describe("Setting", func() {
		Context("when setting to a directory", func() {
			It("should error", func() {
				badMenu := storeadapter.StoreNode{
					Key:   "/menu",
					Value: []byte("oops"),
				}
				err := adapter.SetMulti([]storeadapter.StoreNode{badMenu})
				Ω(err).Should(Equal(storeadapter.ErrorNodeIsDirectory))

				value, err := adapter.Get("/menu/breakfast")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(breakfastNode))
			})
		})

		Context("when implicitly turning a node into a directory", func() {
			It("should error", func() {
				badBreakfast := storeadapter.StoreNode{
					Key:   "/menu/breakfast/elevensies",
					Value: []byte("oops"),
				}
				err := adapter.SetMulti([]storeadapter.StoreNode{badBreakfast})
				Ω(err).Should(Equal(storeadapter.ErrorNodeIsNotDirectory))

				value, err := adapter.Get("/menu/breakfast")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(breakfastNode))
			})
		})

		Context("when overwriting a key", func() {
			It("should overwrite the key", func() {
				discerningBreakfastNode := storeadapter.StoreNode{
					Key:   "/menu/breakfast",
					Value: []byte("crepes"),
				}
				err := adapter.SetMulti([]storeadapter.StoreNode{discerningBreakfastNode})
				Ω(err).ShouldNot(HaveOccurred())

				value, err := adapter.Get("/menu/breakfast")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(discerningBreakfastNode))
			})
		})

		Context("when the key matches the error injector", func() {
			It("should return the injected error", func() {
				lessRandomNode := storeadapter.StoreNode{
					Key:   "/random",
					Value: []byte("0"),
				}

				err := adapter.SetMulti([]storeadapter.StoreNode{lessRandomNode})
				Ω(err).Should(Equal(errors.New("injected set error")))

				adapter.GetErrInjector = nil
				value, err := adapter.Get("/random")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(randomNode))
			})
		})
	})

	Describe("Getting", func() {
		Context("when the key is present", func() {
			It("should return the node", func() {
				value, err := adapter.Get("/menu/breakfast")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(breakfastNode))
			})
		})

		Context("when the key is missing", func() {
			It("should return the key not found error", func() {
				value, err := adapter.Get("/not/a/key")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))
				Ω(value).Should(BeZero())
			})
		})

		Context("when the key is a directory", func() {
			It("should return the key not found error", func() {
				value, err := adapter.Get("/menu")
				Ω(err).Should(Equal(storeadapter.ErrorNodeIsDirectory))
				Ω(value).Should(BeZero())
			})
		})

		Context("when the key matches the error injector", func() {
			It("should return the injected error", func() {
				value, err := adapter.Get("/random")
				Ω(err).Should(Equal(errors.New("injected get error")))
				Ω(value).Should(BeZero())
			})
		})
	})

	Describe("Listing", func() {
		Context("when listing the root directory", func() {
			It("should return the tree of nodes", func() {
				value, err := adapter.ListRecursively("/")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value.Key).Should(Equal("/"))
				Ω(value.Dir).Should(BeTrue())
				Ω(value.ChildNodes).Should(HaveLen(2))
				Ω(value.ChildNodes).Should(ContainElement(randomNode))

				var menuNode storeadapter.StoreNode
				for _, node := range value.ChildNodes {
					if node.Key == "/menu" {
						menuNode = node
					}
				}
				Ω(menuNode.Key).Should(Equal("/menu"))
				Ω(menuNode.Dir).Should(BeTrue())
				Ω(menuNode.ChildNodes).Should(HaveLen(3))
				Ω(menuNode.ChildNodes).Should(ContainElement(breakfastNode))
				Ω(menuNode.ChildNodes).Should(ContainElement(lunchNode))

				var dinnerNode storeadapter.StoreNode
				for _, node := range menuNode.ChildNodes {
					if node.Key == "/menu/dinner" {
						dinnerNode = node
					}
				}
				Ω(dinnerNode.Key).Should(Equal("/menu/dinner"))
				Ω(dinnerNode.Dir).Should(BeTrue())
				Ω(dinnerNode.ChildNodes).Should(HaveLen(2))
				Ω(dinnerNode.ChildNodes).Should(ContainElement(firstCourseDinnerNode))
				Ω(dinnerNode.ChildNodes).Should(ContainElement(secondCourseDinnerNode))
			})
		})

		Context("when listing a subdirectory", func() {
			It("should return the tree of nodes", func() {
				menuNode, err := adapter.ListRecursively("/menu")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(menuNode.Key).Should(Equal("/menu"))
				Ω(menuNode.Dir).Should(BeTrue())
				Ω(menuNode.ChildNodes).Should(HaveLen(3))
				Ω(menuNode.ChildNodes).Should(ContainElement(breakfastNode))
				Ω(menuNode.ChildNodes).Should(ContainElement(lunchNode))

				var dinnerNode storeadapter.StoreNode
				for _, node := range menuNode.ChildNodes {
					if node.Key == "/menu/dinner" {
						dinnerNode = node
					}
				}
				Ω(dinnerNode.Key).Should(Equal("/menu/dinner"))
				Ω(dinnerNode.Dir).Should(BeTrue())
				Ω(dinnerNode.ChildNodes).Should(HaveLen(2))
				Ω(dinnerNode.ChildNodes).Should(ContainElement(firstCourseDinnerNode))
				Ω(dinnerNode.ChildNodes).Should(ContainElement(secondCourseDinnerNode))
			})
		})

		Context("when listing a nonexistent key", func() {
			It("should return the key not found error", func() {
				value, err := adapter.ListRecursively("/not-a-key")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))
				Ω(value).Should(BeZero())
			})
		})

		Context("when listing an entry", func() {
			It("should return the key is not a directory error", func() {
				value, err := adapter.ListRecursively("/menu/breakfast")
				Ω(err).Should(Equal(storeadapter.ErrorNodeIsNotDirectory))
				Ω(value).Should(BeZero())
			})
		})

		Context("when the key matches the error injector", func() {
			It("should return the injected error", func() {
				adapter.ListErrInjector = NewFakeStoreAdapterErrorInjector("menu", errors.New("injected list error"))
				value, err := adapter.ListRecursively("/menu")
				Ω(err).Should(Equal(errors.New("injected list error")))
				Ω(value).Should(BeZero())
			})
		})
	})

	Describe("Deleting", func() {
		Context("when the key is present", func() {
			It("should delete the node", func() {
				err := adapter.Delete("/menu/breakfast", "/menu/lunch")
				Ω(err).ShouldNot(HaveOccurred())

				_, err = adapter.Get("/menu/breakfast")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))

				_, err = adapter.Get("/menu/lunch")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))
			})
		})

		Context("when the key is missing", func() {
			It("should return the key not found error", func() {
				err := adapter.Delete("/not/a/key")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))
			})
		})

		Context("when the key is a directory", func() {
			It("should kaboom the directory", func() {
				err := adapter.Delete("/menu")
				Ω(err).ShouldNot(HaveOccurred())

				_, err = adapter.Get("/menu")
				_, err = adapter.Get("/menu")
				Ω(err).Should(Equal(storeadapter.ErrorKeyNotFound))
			})
		})

		Context("when the key matches the error injector", func() {
			It("should return the injected error", func() {
				err := adapter.Delete("/random")
				Ω(err).Should(Equal(errors.New("injected delete error")))

				value, err := adapter.Get("/menu/breakfast")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(value).Should(Equal(breakfastNode))
			})
		})
	})

	Describe("Watching", func() {
		Context("when a node under the key is created", func() {
			It("sends an event with CreateEvent type and the node's value", func(done Done) {
				events, _, _ := adapter.Watch("/foo")

				err := adapter.Create(storeadapter.StoreNode{
					Key:   "/foo/a",
					Value: []byte("new value"),
				})
				Expect(err).ToNot(HaveOccurred())
				event := <-events

				Expect(event.Type).To(Equal(storeadapter.CreateEvent))
				Expect(event.Node.Key).To(Equal("/foo/a"))
				Expect(string(event.Node.Value)).To(Equal("new value"))

				close(done)
			}, 5.0)
		})

		Context("when a node under the key is updated", func() {
			BeforeEach(func() {
				err := adapter.SetMulti([]storeadapter.StoreNode{
					{
						Key:   "/foo/a",
						Value: []byte("some value"),
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("sends an event with UpdateEvent type and the node's value", func(done Done) {
				events, _, _ := adapter.Watch("/foo")

				err := adapter.SetMulti([]storeadapter.StoreNode{
					{
						Key:   "/foo/a",
						Value: []byte("new value"),
					},
				})
				Expect(err).ToNot(HaveOccurred())

				event := <-events
				Expect(event.Type).To(Equal(storeadapter.UpdateEvent))
				Expect(event.Node.Key).To(Equal("/foo/a"))
				Expect(string(event.Node.Value)).To(Equal("new value"))

				close(done)
			}, 5.0)
		})

		Context("when a node under the key is deleted", func() {
			BeforeEach(func() {
				err := adapter.SetMulti([]storeadapter.StoreNode{
					{
						Key:   "/foo/a",
						Value: []byte("some value"),
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("sends an event with DeleteEvent type and the node's value", func(done Done) {
				events, _, _ := adapter.Watch("/foo")

				err := adapter.Delete("/foo/a")
				Expect(err).ToNot(HaveOccurred())

				event := <-events
				Expect(event.Type).To(Equal(storeadapter.DeleteEvent))
				Expect(event.Node.Key).To(Equal("/foo/a"))
				Expect(string(event.Node.Value)).To(Equal("some value"))

				close(done)
			}, 5.0)
		})

		Context("when a watch error occurs", func() {
			It("returns it to the watcher over the errs channel", func(done Done) {
				_, _, errs := adapter.Watch("/foo")

				disaster := errors.New("oh no!")

				adapter.WatchErrChannel <- disaster

				Expect(<-errs).To(Equal(disaster))

				close(done)
			})
		})
	})
})
