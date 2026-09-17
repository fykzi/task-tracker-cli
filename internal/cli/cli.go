package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fykzi/task-tracker-cli/internal/status"
	"github.com/fykzi/task-tracker-cli/internal/storage"
	"github.com/olekukonko/tablewriter"
)

const (
	minArgsAdd    = 1
	minArgsUpdate = 2
	minArgsDelete = 1
	minArgsMark   = 1
)

type TaskStorage interface {
	Add(description string) (*storage.Task, error)
	Update(id int, description string) (*storage.Task, string, error)
	Delete(id int) (*storage.Task, error)
	ChangeStatus(id int, s status.Status) (*storage.Task, error)
	ListTasks() map[status.Status][]*storage.Task
	ListTasksByStatus(s status.Status) []*storage.Task
}

type CLI struct {
	storage TaskStorage
}
type WrongUsageError struct {
	Message string
	Cause   error
}

type TooFewArgumentsError struct {
	Need int
	Have int
}

func New(s TaskStorage) *CLI {
	return &CLI{storage: s}
}

func (e WrongUsageError) Error() string {
	switch {
	case e.Message != "" && e.Cause != nil:
		return fmt.Sprintf("wrong usage: %s: %s", e.Message, e.Cause.Error())
	case e.Cause != nil:
		return fmt.Sprintf("wrong usage: %s", e.Cause.Error())
	default:
		return fmt.Sprintf("wrong usage: %s", e.Message)
	}
}

func (e TooFewArgumentsError) Error() string {
	return fmt.Sprintf("too few arguments: need %d, have %d", e.Need, e.Have)
}

func (c *CLI) printUsage() {
	fmt.Println(`
USAGE:
	task-tracker-cli [command] [args]...
COMMANDS:
	-h, help                           show help message
	add "<description>"                add new task (description >= 5)
	update <ID> "<new description>"    update task description by id (description >= 5)
	delete <ID>                        delete task by id
	list                               show all tasks
	     todo 		                   show tasks with status "todo"
	     in-progress                   show tasks with status "in-progress"
	     done                          show tasks with status "done"
	mark-in-progress <ID>              change task status with ID on status "in-progress"
	mark-done <ID>                     change task status with ID in status "done"
		`)
}

func (c *CLI) Run(args []string) error {
	if len(args) < 1 {
		return WrongUsageError{Message: "subcommand required"}
	}

	subCommand := args[0]
	subArgs := args[1:]

	switch subCommand {
	case "-h", "help":
		c.printUsage()
		return nil
	case "add":
		return c.addTask(subArgs)
	case "list":
		return c.listTasks(subArgs)
	case "mark-done":
		return c.markTask(subArgs, status.StatusDone)
	case "mark-in-progress":
		return c.markTask(subArgs, status.StatusInProgress)
	case "update":
		return c.updateTask(subArgs)
	case "delete":
		return c.deleteTask(subArgs)
	default:
		return WrongUsageError{
			Message: fmt.Sprintf("unknown subcommand '%s'", subCommand),
		}
	}
}

func (c *CLI) addTask(args []string) error {
	if len(args) < minArgsAdd {
		return WrongUsageError{
			Message: "add",
			Cause: TooFewArgumentsError{
				Need: minArgsAdd,
				Have: len(args),
			},
		}
	}

	description := args[0]
	if description == "" {
		return WrongUsageError{
			Message: "add: invalid description",
		}
	}

	task, err := c.storage.Add(description)
	if err != nil {
		return fmt.Errorf("add task: %w", err)
	}

	fmt.Printf("Task added successfully (ID: %d)\n", task.ID)

	return nil
}

func (c *CLI) updateTask(args []string) error {
	id, rest, err := c.parseID(args, minArgsUpdate, "update")
	if err != nil {
		return err
	}

	newDescription := rest[0]
	if newDescription == "" {
		return WrongUsageError{
			Message: "update: invalid description",
		}
	}

	task, oldDescription, err := c.storage.Update(id, newDescription)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	fmt.Printf("Task(ID: %d) updated from `%s` to '%s'\n", task.ID, oldDescription, task.Description)

	return nil
}

func (c *CLI) deleteTask(args []string) error {
	id, _, err := c.parseID(args, minArgsDelete, "delete")
	if err != nil {
		return err
	}

	task, err := c.storage.Delete(id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	fmt.Printf("Task(ID: %d) deleted successfully\n", task.ID)

	return nil
}

func (c *CLI) markTask(args []string, s status.Status) error {
	id, _, err := c.parseID(args, minArgsMark, "mark")
	if err != nil {
		return err
	}

	task, err := c.storage.ChangeStatus(id, s)
	if err != nil {
		return fmt.Errorf("change status: %w", err)
	}

	fmt.Printf("Change status to `%s` for Task(ID: %d)\n", SprintStatus(s), task.ID)

	return nil
}

func (c *CLI) listTasks(args []string) error {
	var tasks []*storage.Task

	if len(args) == 0 {
		ts := c.storage.ListTasks()

		for _, s := range status.ListStatus {
			tasks = append(tasks, ts[s]...)
		}
	} else {
		if len(args) > 1 {
			return WrongUsageError{Message: "too many arguments for list"}
		}

		st, err := status.Parse(args[0])
		if err != nil {
			return WrongUsageError{
				Message: "list error",
				Cause:   err,
			}
		}

		tasks = c.storage.ListTasksByStatus(st)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Description", "Status", "Created at", "Updated at"})

	for _, t := range tasks {
		appendRowToTable(table, t)
	}

	_ = table.Render()

	return nil
}

func (c *CLI) parseID(args []string, minArgs int, commandName string) (int, []string, error) {
	if len(args) < minArgs {
		return 0, nil, WrongUsageError{
			Message: commandName,
			Cause: TooFewArgumentsError{
				Need: minArgs,
				Have: len(args),
			},
		}
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, nil, WrongUsageError{Message: "invalid ID"}
	}

	return id, args[1:], nil
}

func appendRowToTable(table *tablewriter.Table, t *storage.Task) {
	_ = table.Append(
		strconv.Itoa(t.ID),
		t.Description,
		SprintStatus(t.Status),
		t.CreatedAt.Format("15:04 02.01.2006"),
		t.UpdatedAt.Format("15:04 02.01.2006"),
	)
}
