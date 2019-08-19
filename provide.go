package hub

// Provide is a dependency injection provider, of the form used by packages such as uber/fx.
// This function simply invokes New(), and then returns the same instance as the different available
// interfaces in this package.
//
// Using this function instead of New allows code to more easily consume dependencies with the specific
// interfaces, e.g. Publisher.  For example:
//
//    import "go.uber.org/fx"
//    import "github.com/johnabass/hub"
//
//    func main() {
//        fx.New(
//            fx.Provide(
//                hub.Provide,
//                func(p hub.Publisher) MyComponent {
//                    // use the Publisher
//                },
//                func(s hub.Subscriber) AnotherComponent {
//                    // use the Subscriber
//                },
//            )
//        )
//    }
func Provide() (Publisher, Subscriber, Interface) {
	h := New()
	return h, h, h
}
