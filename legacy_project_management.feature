Feature: Manage a writing project
  So that a writer can continue and shape a manuscript safely
  As a Writer
  I want to save work, reopen prior work, export a draft, and apply a story template

  Scenario: Save the current project with a new name
    Given a Writer has an open project with chapters and notes
    And the project does not yet have a remembered save name
    When the Writer saves the project using a chosen name
    Then the project is stored under that name in the standard project format
    And the system remembers that name for future saves
    And the Writer is told that the project was saved

  Scenario: Save the current project using the remembered name
    Given a Writer has an open project with chapters and notes
    And the project already has a remembered save name
    When the Writer saves the project without entering another name
    Then the project is stored using the remembered name
    And the Writer is told that the project was saved

  Scenario: Refuse to save when no name is available
    Given a Writer has an open project with chapters and notes
    And the project does not have a remembered save name
    When the Writer saves the project without entering a name
    Then the save is refused
    And the Writer is told that a name is required

  Scenario: Open an existing project
    Given a previously saved project exists
    When the Writer opens that project using its name
    Then the project chapters and reference notes are restored
    And the reopened project becomes the current working project
    And the system remembers that name for future saves
    And the Writer is told that the project was loaded

  Scenario: Refuse to open a project when no name is provided
    Given a Writer wants to reopen prior work
    When the Writer opens a project without providing a name
    Then the request is refused
    And the Writer is told that a name is required

  Scenario: Refuse to open a project that cannot be read
    Given a Writer provides the name of a saved project
    And the named project is missing or unreadable
    When the Writer opens that project
    Then the request is refused
    And the current working project remains unchanged

  Scenario: Open an older saved project that does not include reference notes
    Given a previously saved project exists with chapters only
    When the Writer opens that project
    Then the chapters are restored
    And the system creates a general reference notes entry for the Writer
    And the reopened project becomes the current working project

  Scenario: Open a damaged project
    Given a Writer provides the name of a saved project
    And the saved project does not contain usable manuscript content
    When the Writer opens that project
    Then the request is refused
    And the Writer is told that the project is empty or damaged

  Scenario: Export the manuscript as a reading copy
    Given a Writer has an open project with one or more chapters
    When the Writer exports the manuscript using a chosen name
    Then a reading copy is produced under that name in the standard text format
    And each chapter appears with its chapter number and title
    And the Writer is told that the export is complete

  Scenario: Refuse to export when no name is provided
    Given a Writer has an open project with one or more chapters
    When the Writer exports the manuscript without providing a name
    Then the export is refused
    And the Writer is told that a name is required

  Scenario Outline: Apply a supported story template
    Given a Writer has an open project
    When the Writer applies the "<template>" story template
    Then the manuscript is replaced with the chapters for that template
    And the Writer is taken to the first chapter of the new outline
    And the project keeps its reference notes area available
    And the Writer is told that the template was applied

    Examples:
      | template |
      | 3 act    |
      | hero     |
      | save the cat |
      | fichtean |
      | horror   |

  Scenario: Refuse to apply a story template when no template is chosen
    Given a Writer has an open project
    When the Writer asks to apply a story template without naming one
    Then the request is refused
    And the Writer is told to choose a template

  Scenario: Refuse to apply an unknown story template
    Given a Writer has an open project
    When the Writer applies an unrecognized story template
    Then the request is refused
    And the current project remains unchanged
