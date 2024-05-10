(function(){
    angular.module("patio-app")
    .directive("submitButton", function(){
        return {
            restrict: "E",
            scope: {
                submitnow: '&',
                submitdone: "=",
                btntext:"@",
                isoutline:"@" // determines the outline of the button
            },
            templateUrl: "/assets/templates/submit-btn.html",
            controller: function($scope) {
                console.log("indside the submit button" + $scope.btntext);
                $scope.pending = false;
                $scope.outline = ($scope.isoutline == "true");
                $scope.presubmit = function(){
                    $scope.pending = true; 
                    $scope.submitnow();
                    /* do not set this pending=false since submit now has asynch code that return immediately even before the response is received.
                    pending = false is set only revisitng the page and directive is reloaded */
                }
                $scope.$watch('submitdone', function(after, before){
                    /* This directive isnt per say concerned about if the submit was successful / failure it only seeks acknowledgement of the action completion Even if the validation fails and the ACTUAL actin is never executed this would only want to know when it completes */
                    if (after !== undefined && after ==true)  {
                        $scope.pending = false;
                    }
                })
            }
        }
    })
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
                        url: _userauth_baseurl+'?action=auth',
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
    .directive("userDevices", function(){
        return {
            restrict: "EA",
            scope :{},
            templateUrl: "/assets/templates/user-devices.html",
            controller:  function($scope,$window,$http){
                // user id needs to be extracted from the url 
                // userid = $window.localStorage.getItem("user-id");
                // getting the user devices 
                $scope.devices = [
                    {name:"No devices found!", make:"Unfortunately we haven't found any devices that you control. One or more devices when deployed will have your email as owners"}
                ];
                // retreiving user information from local storage 
                // this is required to display the username on the legend
                // also user id is used to form the devices url 
                completeName = $window.localStorage.getItem("user-name");
                if (completeName == "") {
                    $scope.legendTxt = "List of devices"
                } else {
                    names = completeName.split(" ")
                    if (names.length >1) {
                        $scope.legendTxt = names[0]+'\'s devices';
                    } else {
                        $scope.legendTxt = names+'\'s devices';
                    }
                }
                $http ({
                    method:'get',
                    // example url to get all the devices for the user email kneerunjun@gmail.com
                    // http://aqua.eensymachines.in:30001/api/devices?filter=users&user=kneerunjun@gmail.com
                    url: _devicereg_baseurl+'?filter=users&user='+$window.localStorage.getItem("user-email"),
                }).then(function(response){
                    console.log("received data for user devices");
                    // what we receive is an [] of dedvice objects
                    console.log(response.data);
                    $scope.devices  = response.data;
                    $scope.devices.forEach(x => {
                        x.link = {};
                    })
                }, function(response){
                    console.log("error getting the user devices");
                })
            }
        }
    })
})()