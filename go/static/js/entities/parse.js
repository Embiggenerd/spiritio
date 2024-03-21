/**
 * Parses a user message starting with '/' into a work order
 * @type {import("../../types").Parser}
 */
const parser = {
    command: '',
    i: 0,
    workOrderKey: '',
    commandConfigs: {
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
    },
    commandChar: '/',
    allLettersRegex: /[a-zA-Z]/,
    parseUserCommand: function (command) {
        this.command = command
        this.parse()
        let argsCount = 0
        let argsRequired = 0

        this.commandConfigs[this.workOrderKey].args.forEach((a) => {
            if (a.value) {
                argsCount++
            }
            argsRequired++
        })
        if (argsCount !== argsRequired) {
            throw new Error(
                `wrong argument number: have ${argsCount}, want ${argsRequired}`
            )
        }
        return this.commandConfigs[this.workOrderKey]
    },

    parse: function () {
        if (this.match(this.commandChar)) {
            this.eat(this.commandChar)
        }
        this.parseCommand(this.readWhileMatching(this.allLettersRegex))
        this.skipWhitespace()
        this.parseArguments()
    },

    // Find any words that are in the list of possible commands
    parseCommand: function (word) {
        this.workOrderKey = this.workOrderKey + ' ' + word
        this.workOrderKey = this.workOrderKey.trim()
        if (
            Object.keys(this.commandConfigs).includes(
                this.workOrderKey.toLowerCase()
            ) ||
            this.i === this.command.length
        ) {
            return
        }
        this.skipWhitespace()
        this.parseCommand(this.readWhileMatching(this.allLettersRegex))

        // Figure out the arguments according to config
    },
    readWhileMatching: function (regex) {
        let startIndex = this.i
        while (
            this.i < this.command.length &&
            regex.test(this.command[this.i])
        ) {
            this.i++
        }
        return this.command.slice(startIndex, this.i)
    },

    eat: function (str) {
        if (this.match(str)) {
            this.i += str.length
        } else {
            throw new Error(`Parse error: expecting "${str}"`)
        }
    },

    match: function (str) {
        return this.command.slice(this.i, this.i + str.length) === str
    },

    skipWhitespace: function () {
        this.readWhileMatching(/[\s\n]/)
    },
    parseArguments: function () {
        if (!this.commandConfigs.hasOwnProperty(this.workOrderKey)) {
            throw new Error('no such command')
        }
        const args = this.commandConfigs[this.workOrderKey].args
        let j = 0
        if (args) {
            while (j < args.length) {
                this.skipWhitespace()
                args[j].value = this.readWhileMatching(args[j].regex)
                j++
            }
        }
        this.skipWhitespace()
        const next = this.i + 1
        if (this.command[next]) {
            throw new Error('Too many arguments')
        }
    },
}

export default () => {
    console.log('parser called')
    return { ...parser }
}
