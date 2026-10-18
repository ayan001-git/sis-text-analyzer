# Requirement mapping

Source specification: https://github.com/salikhP/GOLANG-FALL-2026-KBTU/blob/0d60b05cf2e270f93203fa2b362dd5b0c0305e8f/sis/SIS_CP1_SPECIFICATION.md

Source guidelines: https://github.com/salikhP/GOLANG-FALL-2026-KBTU/blob/0d60b05cf2e270f93203fa2b362dd5b0c0305e8f/sis/SIS%20Student%20Guidelines.pdf

Reviewed source commit: 0d60b05cf2e270f93203fa2b362dd5b0c0305e8f.

| Requirement | Implementation / evidence |
|---|---|
| Task with ID, input, state, result | Task struct in task/task.go |
| Create, validate, reject invalid input | Service.Create; TestCreateAndGet, TestInvalidInput |
| Unique IDs, memory storage | nextID and map[int]Task; TestIndependentTasks |
| Retrieve and missing ID error | Service.Get; TestCreateAndGet, TestMissingTask |
| Meaningful processing and stored result | TextAnalyzer.Analyze, Service.Process; TestProcessResults |
| Pending, processing, completed | Status constants; TestProcessingState, TestProcessResults |
| Reject invalid operations | Process checks Pending; repeated call tests and processing observer |
| Independent tasks | TestIndependentTasks |
| Structs, methods, pointer/value receivers | Service, Task, TextAnalyzer |
| Meaningful interface | Analyzer; TextAnalyzer and observingAnalyzer implementations |
| Map and explicit errors | Service.tasks; sentinel errors, fmt.Errorf with %w |
| Standard testing and table-driven tests | task/task_test.go; TestProcessResults, TestInvalidInput |
| Valid module, Go 1.27 | go.mod |
| Format, compile, vet, tests | docs/VALIDATION.txt |
| Public GitHub repository and Teams link | Student must publish and submit; not performed by this archive |
| Individual code defense | Student must understand and defend code; docs/DEFENSE_RU_EN.md |

The application is synchronous and uses only memory. No later-checkpoint functionality is included.
