<template>
  <label class="block text-sm">
    <span v-if="label" class="font-medium text-zinc-300">{{ label }}</span>
    <input
      :value="modelValue"
      type="text"
      inputmode="numeric"
      placeholder="YYYY/MM/DD"
      :required="required"
      :class="inputClass"
      @input="$emit('update:modelValue', $event.target.value)"
      @change="$emit('change')"
    />
  </label>
</template>

<script setup>
defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: '' },
  required: { type: Boolean, default: false },
  inputClass: {
    type: String,
    default:
      'mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder:text-zinc-500',
  },
})

defineEmits(['update:modelValue', 'change'])
</script>
