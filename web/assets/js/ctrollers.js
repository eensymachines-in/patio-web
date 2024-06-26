(function () {
    angular.module("patio-app").controller("loginCtrl", function ($scope, $http, $location, $window, $sce) {
        // $scope.loginErr = null;
        $scope.validationErr = false
        var clear_session = function () {
            $window.localStorage.removeItem("user-id");
            $window.localStorage.removeItem("user-email");
            $window.localStorage.removeItem("user-name");
            $window.localStorage.removeItem("user-role");
            $window.localStorage.removeItem("user-telegid");
            $window.localStorage.removeItem("user-authtok");
        }; clear_session();
        $scope.done = false; // this indicates to the directive below that the submit process is complete 
        $scope.movingaway = false;
        $scope.login = {
            usrid: "",
            passwd: "",
            submit: function () {
                // send the login creds back to where this page came from and get the result
                // console.log(this.usrid)
                // console.log(this.passwd)

                if (this.usrid == "" || this.passwd == "") {
                    $scope.validationErr = true;
                    $scope.done = true; // directives would not be concerned if the submit per say is called, only gets to know the action is done
                    return
                }
                $http({
                    method:'post',
                    url: _userauth_baseurl +'?action=auth',
                    data: {email: this.usrid, auth: this.passwd},
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }).then(function (response) {
                    // Instead of the sessionStorage we are using localStorage since when on the same browser we want to carry ahead the login token to a new tab on the same browser 
                    $window.localStorage.setItem("user-id", response.data.id);
                    $window.localStorage.setItem("user-email", response.data.email);
                    $window.localStorage.setItem("user-name", response.data.name);
                    $window.localStorage.setItem("user-role", response.data.role);
                    $window.localStorage.setItem("user-telegid", response.data.telegid);
                    // https://medium.com/kanlanc/heres-why-storing-jwt-in-local-storage-is-a-great-mistake-df01dad90f9e
                    // BUG: look into the article and find out why storing tokens in localstorage isnt a good idea
                    // For now we are just going ahead with the idea of having everything on localstorage 
                    $window.localStorage.setItem("user-authtok", response.data.authtok);
                    $scope.done = true;
                    $location.url("/users/" + response.data.id + "/devices") // device listing page where user can select the devices under his control
                    
                }, function (response) {
                    if (response.status == 401) {
                        // A error modal will indicate the password did not match the records
                        $scope.fMsg =$sce.trustAsHtml("<p>Credentials did not match our records, check and send again</p>");
                    } else if (response.status == 500) {
                        // again showing the error modal with a appropriate message
                        $scope.fMsg =$sce.trustAsHtml("<p>One or more things on the server went wrong.Wait a while before you try again.</p>");
                    } else {
                        $scope.fMsg =$sce.trustAsHtml("<p>Something on the server went wrong and we have no idea what.</p> <p>Time to call an administrator</p>");
                    }
                    $scope.done = true;
                    $("#failModal").modal('show'); // finally the error modal comes up
                })
            }
        }
        $scope.gotoSingup = function(){
            $scope.movingaway = true;
            // $location.url("/signup");
        }
    })
        .controller("settingsCtrl", function ($scope, $http, $timeout, $routeParams, $sce) {
            $scope.sMsg = ""; // success message for the modal when the operation is success 
            $scope.hrOptions = [
                "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
                "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"
            ];
            $scope.mnOptions = [
                "00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
                "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25",
                "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38",
                "39", "40", "41", "42", "43", "45", "46", "47", "48", "49", "50", "51", "52",
                "53", "54", "55", "56", "57", "58", "59"
            ]
            $scope.configOptions = [
                { opt: 0, txt: "Tick Every Interval", note: "One trigger after every interval, time of the day is irrelevant.", exmp: "Example: ON for 50 seconds, followed by OFF for 50 seconds (interval) in an infinite cycle." },
                { opt: 1, txt: "Tick Every Day At", note: "One trigger at specific time of day, interval is irrelevant.", exmp: "Example: ON at 10:00 one day, OFF the next day at 10:00 in an infinite cycle." },
                { opt: 2, txt: "Pulse Every Interval", note: "Two triggers separated by pulse gap after every interval.", exmp: "Example: ON for 50 seconds (pulse), OFF for 30 seconds, repeat after every 80 seconds (interval) in an infinite cycle." },
                { opt: 3, txt: "Pulse Every Day At", note: "Two triggers separated by pulse gap at specific time of the day", exmp: "Example: ON for 50 seconds, (pulse) OFF after that every day at 10:00, repeat for each day." }
            ]
            $scope.viewModel = { // the model that we used to communicate to the view 
                // viewModel is populated with dedfault values, but when the settings are downloaded this will actual current values
                clock: null,
                CfgOpt: $scope.configOptions[0],
                pulsegap: 60,
                interval: 100,
                // for gui purpose 
                pulsegapInvalid: false,
                intervalInvalid: false
            }
            $scope.settings = { //the model that shall be dispatched away from the controller mostly to the server
                // pump settings 
                config: 0,
                tickat: "",
                pulsegap: 60,
                interval: 100,
            }
            $scope.$watch("viewModel.pulsegap", function (after, before) {
                if (after <= 0 || after > 86340) {
                    $scope.viewModel.pulsegapInvalid = true;
                    return
                }
                // cannot have a pulse gap which is 0 
                if ($scope.viewModel.CfgOpt.opt == 2) {
                    if (after >= $scope.viewModel.interval) {
                        $scope.viewModel.pulsegapInvalid = true;
                        return
                    }
                }
                $scope.viewModel.pulsegapInvalid = false;
                $scope.settings.pulsegap = after
            })
            $scope.$watch("viewModel.interval", function (after, before) {
                if (after <= 0 || after > 86400) {
                    console.log("out of bounds interval")
                    $scope.viewModel.intervalInvalid = true;
                    return
                }
                if ($scope.viewModel.CfgOpt.opt == 2) {
                    if (after <= $scope.viewModel.pulsegap) {
                        console.log("interval cannot be equal to pulse gap")
                        $scope.viewModel.intervalInvalid = true;
                        return
                    }
                }
                $scope.viewModel.intervalInvalid = false;
                $scope.settings.interval = after
            })
            $scope.$watch("viewModel.CfgOpt", function (after, before) {
                if (after !== undefined && after !== null) {
                    $scope.settings.config = after.opt;

                }
            })
            $scope.$watch("viewModel.clock", function (after, before) {
                if (after !== undefined && after !== null) {
                    $scope.settings.tickat = after.hr + ":" + after.min;
                }
            }, true) // its a deep watch since we want to track the properties
            $scope.updateDone = false; // as an indicator that changes have been submitted to the device
            /* Irrespective of whether its a success response or not this will basically iindicate that the http action is done */
            $scope.submit = function () {
                console.log($scope.settings);
                $http({
                    method: 'patch',
                    url: _devicereg_baseurl +'/'+ $routeParams.deviceID + '?path=config&action=replace',
                    data: $scope.settings,
                    headers: {
                        'Content-Type': "application/json"
                    },
                }).then(function (response) {
                    console.log("done! settings have been updated", response);
                    $scope.updateDone = true;
                    $scope.sMsg = $sce.trustAsHtml("<p>Device configuration is modified!<br>If the device is online changes would reflect instanteneous</p>");
                    $("#successModal").modal('show');
                    // $route.reload(); // reload the same settings page 
                }, function (response) {
                    $scope.updateDone = true;
                    if (response.status == 400) {
                        console.error("Bad request updating the device configuration" + response);
                    }
                    console.error(response + "failed ! settings could not be updated");
                    $scope.fMsg =$sce.trustAsHtml("Device configuration could not be updated. Contact the administrator to know more");
                    $("#failModal").modal('show'); // finally the error modal comes up
                })
            }
            // Getting the current settings to start with 
            $http({
                method: 'get',
                url: _devicereg_baseurl +'/'+ $routeParams.deviceID ,
            }).then(function (response) {
                console.log("received current settings from the server", response.data);
                $scope.configOptions.forEach(e => {
                    if (e.opt == response.data.cfg.config) {
                        console.log("found matching config option..")
                        $scope.viewModel.CfgOpt = e;
                        return
                    }
                })
                $scope.viewModel.clock = { hr: response.data.cfg.tickat.split(":")[0], min: response.data.cfg.tickat.split(":")[1] };
                $scope.viewModel.pulsegap = response.data.cfg.pulsegap;
                $scope.viewModel.interval = response.data.cfg.interval;

                $timeout(function () {
                    console.log("viewmodel to settings ...");
                    console.log($scope.settings);
                }, 500)
            }, function (response) {
                if (response.status == 404) {
                    $scope.fMsg =$sce.trustAsHtml("<p>Device you were looking for is not found registered on our records</p>");
                } else if (response.status == 500) {
                    $scope.fMsg =$sce.trustAsHtml("<p>One or more things on the server went wrong.Wait a while before you try again.</p>");
                } else if (response.status == 400) {
                    $scope.fMsg =$sce.trustAsHtml("<p>Unexpected request params caused this. Time to call an administrator</p>");
                }
                $("#failModal").modal('show'); // finally the error modal comes up
            })

        })
})();