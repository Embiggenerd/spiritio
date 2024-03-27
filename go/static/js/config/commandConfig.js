const commandConfigs = {
    'set password': {
        workOrder: 'set_user_password',
        args: [
            {
                name: 'password',
                regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                value: '',
            },
        ],
    },
    'set name': {
        workOrder: 'set_user_name',
        args: [
            {
                name: 'name',
                regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                value: '',
            },
        ],
    },
    login: {
        workOrder: 'validate_user_name_password',
        args: [
            // order of these matters
            {
                name: 'name',
                regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                value: '',
            },
            {
                name: 'password',
                regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                value: '',
            },
        ],
    },
    'direct message': {
        aliases: ['@'],
        workOrder: 'user_message',
        args: [
            {
                name: 'toUserID',
                regex: /[A-Za-z0-9_@./#&+!$_*+-]/,
                value: '',
            },
            {
                name: 'text',
                regex: /./,
                value: '',
            },
        ],
    },
}

export default commandConfigs
