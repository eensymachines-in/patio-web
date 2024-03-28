(function(){
    angular.module("patio-app")
    .directive("authorizedUser", function(){
        /*A simple attribute level directive that when inserted on the top level of a view can make sure to 
        check the token validity with the server and if invalid can redirect the page to login
        For this to work correctly the user shall have signed in already and the token needs to be in the session. 
        This directive expects to have the token in the localStorage 
        */ 
        return {
            restrict: 'A',
            scope: true, 
            controller: function($scope,$window, $location, $http){
               console.log("authorizing the user ..");
               tok = $window.localStorage.getItem("user-authtok");
               if (tok === null) {
                    console.log("Aint no token found, going back to login");
                    $location.url("/"); // token isnt found, we are going back
               } else {
                    $http({
                        method: 'get',
                        url: '/api/authorize',
                        headers: {
                            "Authorization": tok
                        },
                    }).then(function(response){
                        console.log("authorization success ..");
                    }, function(response){
                        console.error("failed authorization..", response.status);
                        $location.url("/");
                    })
               }
            }
        }
    })
})()