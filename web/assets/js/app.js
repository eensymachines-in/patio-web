(function() {
    // Define the angular application 
    angular.module("patio-app", ["ngRoute"]).config(function($interpolateProvider, $locationProvider, $routeProvider, $provide) {
        $interpolateProvider.startSymbol("{[")
        $interpolateProvider.endSymbol("]}")
        $locationProvider.html5Mode({
            enabled: true,
            requireBase: true
        });
        // Add more when clauses to add more views 
        $routeProvider
            .when("/", {
                templateUrl: "/assets/views/splash.html"
            }).when("/settings", {
                templateUrl: "/assets/views/settings.html"
            });

        $provide.provider("emailPattern", function() {
            this.$get = function() {
                // [\w] is the same as [A-Za-z0-9_-]
                // 3 groups , id, provider , domain also a '.' in between separated by @
                // we are enforcing a valid email id 
                // email id can have .,_,- in it and nothing more 
                return /^[\w-._]+@[\w]+\.[a-z]+$/
            }
        })
        $provide.provider("passwdPattern", function() {
            this.$get = function() {
                // here for the password the special characters that are not allowed are being singled out and denied.
                // apart form this all the characters will be allowed
                // password also has a restriction on the number of characters in there
                return /^[\w-!@#%&?_]{8,16}$/
            }
        })
    })
})()