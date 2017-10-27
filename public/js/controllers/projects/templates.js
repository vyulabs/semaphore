define(['controllers/projects/taskRunner'], function () {
	app.registerController('ProjectTemplatesCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', '$window', function ($scope, $http, $modal, Project, $rootScope, $window) {
		$http.get(Project.getURL() + '/keys?type=ssh').success(function (keys) {
			$scope.sshKeys = keys;

			$scope.sshKeysAssoc = {};
			keys.forEach(function (k) {
				if (k.removed) k.name = '[removed] - ' + k.name;
				$scope.sshKeysAssoc[k.id] = k;
			});
		});
		$http.get(Project.getURL() + '/inventory').success(function (inv) {
			$scope.inventory = inv;

			$scope.inventoryAssoc = {};
			inv.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;
				$scope.inventoryAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/repositories').success(function (repos) {
			$scope.repos = repos;

			$scope.reposAssoc = {};
			repos.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

				$scope.reposAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/environment').success(function (env) {
			$scope.environment = env;

			$scope.environmentAssoc = {};
			env.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

				$scope.environmentAssoc[i.id] = i;
			});
		});

		function getHiddenTemplates() {
			try {
				return JSON.parse($window.localStorage.getItem('hidden-templates') || '[]');
			} catch(e) {
				return [];
			}
		}

		function setHiddenTemplates(hiddenTemplates) {
			$window.localStorage.setItem('hidden-templates', JSON.stringify(hiddenTemplates));
		}

		$scope.hasHiddenTemplates = function() {
			return getHiddenTemplates().length > 0;
		};

		function hashCode(str) {
			var hash = 0;
			for (var i = 0; i < str.length; i++) {
				hash = str.charCodeAt(i) + ((hash << 5) - hash);
			}
			return hash;
		}

		function intToRGB(i){
			var c = (i & 0x00FFFFFF).toString(16).toUpperCase();
			return "0000000".substring(0, 6 - c.length) + c;
		}

		function hexToRGBA(hex){
		  var c;
		  if(/^#([A-Fa-f0-9]{3}){1,2}$/.test(hex)){
			c= hex.substring(1).split('');
			if(c.length === 3){
			  c= [c[0], c[0], c[1], c[1], c[2], c[2]];
			}
			c= '0x'+c.join('');
			return 'rgba('+[(c>>16)&255, (c>>8)&255, c&255].join(',')+',0.07)';
		  }
		  throw new Error('Bad Hex');
		}

		function parseAlias(str) {
			var spaceIndex = str.indexOf(' ');
			var path = spaceIndex === -1 ? str : str.substring(0, spaceIndex);
			var pathParts = path.split('/');
			var opts = spaceIndex === -1 ? null : str.substring(spaceIndex + 1);
			return {
				env: pathParts[0],
					role: pathParts[1] || '',
					inventory: pathParts[2] || '',
					options: opts ? opts.split(' ') : []
				};
			}

		$scope.reload = function () {
			$http.get(Project.getURL() + '/templates?sort=alias&order=asc').success(function (templates) {
				var hiddenTemplates = getHiddenTemplates();
				var prev;
				for (var i = 0; i < templates.length; i++) {
					var template = templates[i];
					if (hiddenTemplates.indexOf(template.id) !== -1) {
						template.hidden = true;
					}

					var alias = parseAlias(template.alias);
					if (prev && prev.env !== alias.env) {
						template.isNewEnvironment = true;
					}

					template.backgroundColor = hexToRGBA('#' + intToRGB(hashCode('/' + alias.role)));

					prev = alias;
				}
				$scope.templates = templates;
			});
		}

		$scope.remove = function (template) {
			$http.delete(Project.getURL() + '/templates/' + template.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete template..', 'error');
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;
			scope.buildTemplates = $scope.templates.filter(function(template) { return template.type === 'build'; });

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});
		}

		$scope.update = function (template) {
			var scope = $rootScope.$new();
			scope.tpl = template;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;
			scope.buildTemplates = $scope.templates.filter(function(template) { return template.type === 'build'; });

			var modal = $modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			});

			modal.result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(template);
				}

				var tpl = opts.template;
				$http.put(Project.getURL() + '/templates/' + template.id, tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});

			modal.closed.then(function () {
				$scope.reload();
			});
		}

		$scope.run = function (tpl) {
			$modal.open({
				templateUrl: '/tpl/projects/createTaskModal.html',
				controller: 'CreateTaskCtrl',
				resolve: {
					Project: function () {
						return Project;
					},
					Template: function () {
						return tpl;
					}
				}
			}).result.then(function (task) {
				var scope = $rootScope.$new();
				scope.task = task;
				scope.project = Project;

				$modal.open({
					templateUrl: '/tpl/projects/taskModal.html',
					controller: 'TaskCtrl',
					scope: scope,
					size: 'lg'
				});
			});
		}

		$scope.showAll = function() {
			$scope.allShown = true;
		}

		$scope.hideHidden = function() {
			$scope.allShown = false;
		}

		$scope.hideTemplate = function(template) {
			var hiddenTemplates = getHiddenTemplates();
			if (hiddenTemplates.indexOf(template.id) === -1) {
				hiddenTemplates.push(template.id);
			}
			setHiddenTemplates(hiddenTemplates);
			template.hidden = true;
		}

		$scope.showTemplate = function(template) {
			var hiddenTemplates = getHiddenTemplates();
			var i = hiddenTemplates.indexOf(template.id);
			if (i !== -1) {
				hiddenTemplates.splice(i, 1);
			}
			setHiddenTemplates(hiddenTemplates);
			delete template.hidden;
		}

		$scope.showTasks = function(template) {
			var scope = $rootScope.$new();
			$modal.open({
				templateUrl: '/tpl/projects/templateTasksModal.html',
				scope: scope,
				controller: ['Project', 'Template', function(Project, Template) {
					$http.get(Project.getURL() + '/templates/' + Template.id + '/tasks/last').then(function(tasks) {
						scope.tasks = tasks.data || [];
					});
				}],
				resolve: {
					Project: function () {
						return Project;
					},
					Template: function () {
						return template;
					}
				}
			});
		}

		$scope.copy = function (template) {
			var tpl = angular.copy(template);
			tpl.id = null;

			var scope = $rootScope.$new();
			scope.tpl = tpl;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;
			scope.buildTemplates = $scope.templates.filter(function(template) { return template.type === 'build'; });

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).success(function () {
					$scope.reload();
				}).error(function (_, status) {
					swal('error', 'could not add template:' + status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});