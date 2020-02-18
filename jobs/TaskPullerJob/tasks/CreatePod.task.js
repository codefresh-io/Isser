const Joi = require('joi');
const Base = require('../../BaseJob');
const _ = require('lodash');
const { TASK_PRIORITY } = require('../../../constants');

const ERROR_MESSAGES = {
	FAILED_TO_EXECUTE_TASK: 'Failed to run task CreatePod',
};

class CreatePodTask extends Base {
	async run(task) {
		this.logger.info('Running CreatePod task');
		try {
			const service = await this.getKubernetesService(_.get(task, 'metadata.reName'));
			const pod = await service.createPod(this.logger, task.spec);
			return pod;
		} catch (err) {
			const message = `${ERROR_MESSAGES.FAILED_TO_EXECUTE_TASK}: ${err.message}`;
			this.logger.error(message);
			throw new Error(message);
		}
	}

	async validate(task) {
		return Joi.validate(task, CreatePodTask.validationSchema);
	}
}

CreatePodTask.priority = TASK_PRIORITY.HIGH;
CreatePodTask.Errors = ERROR_MESSAGES;
CreatePodTask.validationSchema = Joi.object().keys({
	spec: Joi.object().required(),
}).options({ stripUnknown: true });
module.exports       = CreatePodTask;
