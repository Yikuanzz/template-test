module.exports = {
    extends: ['@commitlint/config-conventional'],
    rules: {
        'type-enum': [2, 'always', [
            'fix',
            'feat',
            'docs',
            'style',
            'refactor',
            'perf',
            'test',
            'revert',
            'chore',
            'build',
            'ci'
        ]]
    },
    parserPreset: {
        parserOpts: {
            issuePrefixes: ['#']
        }
    },
    helpUrl: 'https://github.com/conventional-changelog/commitlint/#what-is-commitlint',
};