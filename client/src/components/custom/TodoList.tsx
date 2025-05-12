import { Flex, Spinner, Stack, Text } from "@chakra-ui/react";
import TodoItem from "./TodoItem";
import { useQuery } from "@tanstack/react-query";
import { BASE_URL } from "../../App";

export type Todo = {
  _id: number;
  body: string;
  completed: boolean;
};

const TodoList = () => {
  const { data: todos, isLoading } = useQuery<Todo[]>({
    queryKey: ["todos"],
    queryFn: async () => {
      try {
        const res = await fetch(`${BASE_URL}/todos`);
        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Error fetching todo data");
        }

        return data || [];
      } catch (error) {
        console.log(error);
      }
    },
  });

  return (
    <>
      <Text
        fontSize={"4xl"}
        textTransform={"uppercase"}
        fontWeight={"bold"}
        textAlign={"center"}
        my={2}
      >
        Today's Tasks
      </Text>
      {isLoading && (
        <Flex justifyContent={"center"} my={4}>
          <Spinner size={"xl"} />
        </Flex>
      )}
      {!isLoading && todos?.length === 0 && (
        <Stack alignItems={"center"} gap="3">
          <Text fontSize={"xl"} textAlign={"center"} color={"gray.500"}>
            All tasks completed! 🤞
          </Text>
        </Stack>
      )}
      <Stack gap={3} border={"5"}>
        {todos
          ?.sort((a, b) => Number(b.completed) - Number(a.completed))
          .map((todo) => (
            <TodoItem key={todo._id} todo={todo} />
          ))}
      </Stack>
    </>
  );
};
export default TodoList;
